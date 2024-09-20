package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg"
	"github.com/m41denx/particle-engine/pkg/layer"
	"github.com/m41denx/particle-engine/pkg/manifest"
	"github.com/m41denx/particle-engine/pkg/runner"
	"github.com/m41denx/particle-engine/structs"
	"github.com/m41denx/particle-engine/utils"
	"github.com/m41denx/particle-engine/utils/downloader"
	cp "github.com/otiai10/copy"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type BuildContext struct {
	Manifest          manifest.Manifest
	config            *structs.Config
	longrecipe        []*RecipeWorker
	layers            map[string]*layer.Layer
	touchedAppliances bool
	integrityData     map[string]string
	ldir              string
	homedir           string
	builddir          string
	runnerType        string
	runnerInstance    runner.Runner
}

func NewBuildContext(manif manifest.Manifest, ldir string, config *structs.Config) *BuildContext {
	home, _ := os.UserCacheDir()
	pc := filepath.Join(home, "particle_cache")
	manifestForHash := manif
	manifestForHash.Layer.Block = "[sha256 autogen]"
	manifestForHash.Runnable = manifest.RunnableStanza{}
	bdir := filepath.Join(pc, "temp", utils.CalcHash([]byte(manifestForHash.ToYaml())))
	_ = os.MkdirAll(bdir, 0750)
	return &BuildContext{
		Manifest:          manif,
		config:            config,
		runnerType:        "thin",
		layers:            make(map[string]*layer.Layer),
		integrityData:     make(map[string]string),
		longrecipe:        make([]*RecipeWorker, 0),
		touchedAppliances: false,
		ldir:              ldir,
		homedir:           pc,
		builddir:          bdir,
	}
}

func (ctx *BuildContext) Export() error {
	fmt.Println(color.BlueString("Exporting build %s to %s", filepath.Base(ctx.builddir)[:12], ctx.ldir))
	return cp.Copy(filepath.Join(ctx.builddir, "build"), ctx.ldir)
}

func (ctx *BuildContext) Clean(all bool) error {
	if all {
		files, err := os.ReadDir(filepath.Clean(filepath.Join(ctx.homedir, "temp")))
		if err != nil {
			return err
		}
		for _, file := range files {
			if !file.IsDir() {
				continue
			}
			if err := ctx.cleanCache(file.Name()); err != nil {
				return err
			}
		}
	}
	return ctx.cleanCache(filepath.Base(ctx.builddir))
}

func (ctx *BuildContext) cleanCache(id string) error {
	fmt.Println(color.CyanString("Cleaning cache for Manifest %s...", id[:12]))
	return os.RemoveAll(filepath.Join(filepath.Dir(ctx.builddir), id))
}

func (ctx *BuildContext) FetchDependencies() error {
	headWorker := NewRecipeWorker(ctx, nil, ctx.Manifest)
	prg := utils.NewTreeProgress()
	err := prg.TrackFunction(color.CyanString("Fetching dependencies..."), headWorker.fetchChildren)
	if err != nil {
		return err
	}
	fmt.Println(color.BlueString(
		"Need to download %d layers for Manifest %s",
		len(ctx.layers), filepath.Base(ctx.builddir)[:12],
	))
	return nil
}

func (ctx *BuildContext) PrepareEnvironment() error {
	prg := utils.NewTreeProgress()
	if err := ctx.fetchRunner(); err != nil {
		return err
	}

	if err := ctx.DownloadLayers(); err != nil {
		return err
	}

	// Prepare environment
	for i, worker := range ctx.longrecipe {
		if len(worker.manifest.Runnable.Build) > 0 && !ctx.touchedAppliances {
			ctx.touchedAppliances = true
			if err := prg.TrackFunction(
				color.BlueString("Calculating integrity hashes..."), ctx.saveIntegrityHash,
			); err != nil {
				return err
			}
		}
		err := prg.TrackFunction(color.BlueString("Extracting %s...", worker.manifest.Name), func() error {
			return worker.ExtractLayer(ctx.builddir, i == 0)
		})
		if err != nil {
			return err
		}
		if i == 0 {
			if err := ctx.installRunner(); err != nil {
				return err
			}
			continue
		}
		// FIXME: WHERE THE FUCK ARE OVERRIDES AND HOW ARE WE GONNA SHOVE THEM INTO THE ENVIRONMENT
		if len(worker.manifest.Runnable.Build) == 0 {
			continue
		}
		err = prg.TrackFunction(color.BlueString("Running %s...", worker.manifest.Name), func() error {
			return worker.RunAppliance(ctx.builddir, manifest.RecipeLayerStanza{})
		})
		if err != nil {
			return err
		}
	}
	if len(ctx.integrityData) == 0 {
		if err := prg.TrackFunction(
			color.BlueString("Calculating integrity hashes..."), ctx.saveIntegrityHash,
		); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *BuildContext) DownloadLayers() error {
	mgr := downloader.NewDownloader(pkg.NUMCPU)
	mgr.ShowBarAuto()

	for _, l := range ctx.layers {
		err := l.Download(mgr)
		if err != nil {
			return err
		}
	}
	errs := mgr.Do()
	if len(errs) > 0 {
		for i, err := range errs {
			fmt.Println(color.RedString("\nError %d: ", i+1), err)
		}
		return errors.New("multiple errors occurred while pulling layers")
	}
	return nil
}

func (ctx *BuildContext) Build() error {
	fmt.Println(color.CyanString("Building particle %s...", ctx.Manifest.Name))
	prg := utils.NewTreeProgress()
	fmt.Println(color.BlueString("Verifying contents..."))
	prg.Tab()
	if err := ctx.buildDiff(); err != nil {
		return err
	}
	prg.Ret()
	if err := prg.TrackFunction(color.BlueString("Building layer..."), ctx.makeLayer); err != nil {
		return err
	}
	ctx.Manifest.SaveTo(filepath.Join(ctx.ldir, "particle.yaml"))
	return nil
}

func (ctx *BuildContext) Enter() error {
	_, err := os.Stat(filepath.Join(ctx.builddir, "msys2.exe"))
	if err != nil {
		_, err = os.Stat(filepath.Join(ctx.builddir, "bin", "arch-chroot"))
	}
	if err != nil {
		fmt.Println(color.CyanString("Entering system environment..."))
		ctx.runnerInstance = runner.NewThinRunner(ctx.builddir)
		return ctx.runnerInstance.Run("bash", nil)
	}
	fmt.Println(color.CyanString("Entering Full Arch environment..."))
	ctx.runnerInstance = runner.NewFullRunner(ctx.builddir)
	return ctx.runnerInstance.Run("bash", nil)
}

func (ctx *BuildContext) makeLayer() error {
	buildcacheDir := filepath.Join(ctx.builddir, "tmp", "buildcache")
	l, err := layer.CreateLayerFrom(buildcacheDir, layer.NewLayer("", ctx.homedir, ""))
	if err != nil {
		return err
	}
	if err := os.RemoveAll(buildcacheDir); err != nil {
		return err
	}
	ctx.Manifest.Layer.Block = l.Hash
	return nil
}

func (ctx *BuildContext) fetchRunner() error {
	if ctx.runnerType == "full" {
		ctx.runnerInstance = runner.NewFullRunner(ctx.builddir)
	} else {
		ctx.runnerInstance = runner.NewThinRunner(ctx.builddir)
	}
	meta, err := manifest.ParseParticleURL(ctx.runnerInstance.GetDependencyString())
	if err != nil {
		return err
	}
	worker, err := NewRecipeWorkerFromURL(ctx, nil, meta)
	if err != nil {
		return err
	}
	if worker.manifest.Name != "blank" {
		l := layer.NewLayer(worker.manifest.Layer.Block, ctx.homedir, worker.manifest.Layer.Server)
		ctx.hookAddLayer(l)
		worker.layer = l
	}
	ctx.longrecipe = slices.Concat([]*RecipeWorker{worker}, ctx.longrecipe)
	return nil
}

func (ctx *BuildContext) calculateIntegrityHash() error {
	d := filepath.Join(ctx.builddir, "build")
	files := utils.LDir(d, "")

	for _, f := range files {
		h, err := utils.CalcFileHash(filepath.Join(d, f))
		if err != nil {
			return err
		}
		ctx.integrityData[f] = h
	}
	return nil
}

func (ctx *BuildContext) saveIntegrityHash() error {
	if err := ctx.calculateIntegrityHash(); err != nil {
		return err
	}
	integr, err := json.MarshalIndent(ctx.integrityData, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ctx.builddir, "integrity.json"), integr, 0755)
}

func (ctx *BuildContext) buildDiff() error {
	buildcacheDir := filepath.Join(ctx.builddir, "tmp", "buildcache")
	if err := os.RemoveAll(buildcacheDir); err != nil {
		return err
	}
	prg := utils.NewTreeProgress()
	oldHashes := make(map[string]string)
	deletions := make([]string, 0)
	d, err := os.ReadFile(filepath.Join(ctx.builddir, "integrity.json"))
	if err1 := json.Unmarshal(d, &oldHashes); err != nil || err1 != nil {
		return errors.New("unable to read integrity.json: " + errors.Join(err, err1).Error())
	}

	if err := prg.TrackFunction(
		color.BlueString("Calculating integrity hashes..."), ctx.calculateIntegrityHash,
	); err != nil {
		return err
	}

	for file, newHash := range ctx.integrityData {
		oldHash, ok := oldHashes[file]
		if ok && newHash == oldHash {
			continue
		}
		fmt.Println(color.GreenString("+ %s [%s]", file, newHash[:12]))
		_ = os.MkdirAll(filepath.Dir(filepath.Join(buildcacheDir, file)), 0755)
		fh, err1 := os.Open(filepath.Join(ctx.builddir, "build", file))
		nh, err2 := os.Create(filepath.Join(buildcacheDir, file))
		if err1 != nil || err2 != nil {
			return errors.Join(err1, err2)
		}

		_, err3 := io.Copy(nh, fh)
		_ = fh.Close()
		_ = nh.Close()
		if err3 != nil {
			return err3
		}
	}

	for file, oldHash := range oldHashes {
		_, ok := ctx.integrityData[file]
		if ok {
			continue
		}
		fmt.Println(color.RedString("- %s [%s]", file, oldHash[:12]))
		deletions = append(deletions, file)
	}

	if len(deletions) > 0 || true { // FIXME: For whatever forsaken reason, if no files in folder exists, then archive won't build
		deletionsData := strings.Join(deletions, "\n")
		if err = os.WriteFile(
			filepath.Join(buildcacheDir, ".deletions"), []byte(deletionsData), 0755,
		); err != nil {
			return err
		}
	}

	return nil
}

func (ctx *BuildContext) installRunner() error {
	return ctx.runnerInstance.CreateEnvironment()
}

func (ctx *BuildContext) hookAddLayer(layer *layer.Layer) {
	ctx.layers[layer.Hash] = layer
}

func (ctx *BuildContext) hookPushRecipe(rw *RecipeWorker) {
	ctx.longrecipe = append(ctx.longrecipe, rw)
}

func (ctx *BuildContext) GetHomeDir() string {
	return ctx.homedir
}
