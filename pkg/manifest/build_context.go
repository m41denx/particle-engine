package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg"
	"github.com/m41denx/particle-engine/pkg/layer"
	"github.com/m41denx/particle-engine/pkg/runner"
	"github.com/m41denx/particle-engine/structs"
	"github.com/m41denx/particle-engine/utils"
	"github.com/m41denx/particle-engine/utils/downloader"
	"os"
	"path"
	"path/filepath"
	"slices"
)

type BuildContext struct {
	Manifest          Manifest
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

func NewBuildContext(manifest Manifest, ldir string, config *structs.Config) *BuildContext {
	home, _ := os.UserCacheDir()
	pc := filepath.Join(home, "particle_cache")
	bdir := path.Join(pc, "temp", utils.CalcHash([]byte(manifest.ToYaml())))
	os.MkdirAll(bdir, 0750)
	return &BuildContext{
		Manifest:          manifest,
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

func (ctx *BuildContext) FetchDependencies() error {
	headWorker := NewRecipeWorker(ctx, nil, ctx.Manifest)
	if err := headWorker.fetchChildren(); err != nil {
		return err
	}
	fmt.Printf("%#v", ctx)
	return nil
}

func (ctx *BuildContext) PrepareEnvironment() error {
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
			if err := ctx.calculateIntegrityHash(); err != nil {
				return err
			}
		}
		if err := worker.ExtractLayer(ctx.builddir, i == 0); err != nil {
			return err
		}
		if i == 0 {
			if err := ctx.installRunner(); err != nil {
				return err
			}
			continue
		}
		// FIXME: WHERE THE FUCK ARE OVERRIDES AND HOW ARE WE GONNA SHOVE THEM INTO THE ENVIRONMENT
		if err := worker.RunAppliance(ctx.builddir, RecipeLayerStanza{}); err != nil {
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

func (ctx *BuildContext) fetchRunner() error {
	if ctx.runnerType == "full" {
		ctx.runnerInstance = runner.NewFullRunner(ctx.builddir)
	} else {
		ctx.runnerInstance = runner.NewThinRunner(ctx.builddir)
	}
	meta, err := ParseParticleURL(ctx.runnerInstance.GetDependencyString())
	if err != nil {
		return err
	}
	worker, err := NewRecipeWorkerFromURL(ctx, nil, meta)
	if err != nil {
		return err
	}
	l := layer.NewLayer(worker.manifest.Layer.Block, ctx.homedir, worker.manifest.Layer.Server)
	ctx.hookAddLayer(l)
	ctx.longrecipe = slices.Concat([]*RecipeWorker{worker}, ctx.longrecipe)
	return nil
}

func (ctx *BuildContext) calculateIntegrityHash() error {
	d := path.Join(ctx.builddir, "build")
	files := utils.LDir(d, "")

	for _, f := range files {
		h, err := utils.CalcFileHash(path.Join(d, f))
		if err != nil {
			return err
		}
		ctx.integrityData[f] = h
	}

	integr, err := json.MarshalIndent(ctx.integrityData, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(ctx.builddir, "integrity.json"), integr, 0755)
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
