package manifest

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg"
	"github.com/m41denx/particle-engine/pkg/layer"
	"github.com/m41denx/particle-engine/pkg/runner"
	"github.com/m41denx/particle-engine/structs"
	"github.com/m41denx/particle-engine/utils/downloader"
	"os"
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
	runnerType        string
	runnerInstance    runner.Runner
}

func NewBuildContext(manifest Manifest, ldir string, config *structs.Config) *BuildContext {
	home, _ := os.UserCacheDir()
	pc := filepath.Join(home, "particle_cache")
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

func (ctx *BuildContext) downloadLayers() error {
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

func (ctx *BuildContext) installRunner() error {
	if ctx.runnerType == "full" {
		ctx.runnerInstance = runner.NewFullRunner(ctx.ldir)
	} else {
		ctx.runnerInstance = runner.NewThinRunner(ctx.ldir)
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

func (ctx *BuildContext) hookAddLayer(layer *layer.Layer) {
	ctx.layers[layer.Hash] = layer
}

func (ctx *BuildContext) hookPushRecipe(rw *RecipeWorker) {
	ctx.longrecipe = append(ctx.longrecipe, rw)
}
