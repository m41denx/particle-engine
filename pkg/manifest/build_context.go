package manifest

import (
	"fmt"
	"github.com/m41denx/particle-engine/pkg/layer"
	"github.com/m41denx/particle-engine/structs"
	"os"
	"path/filepath"
)

type BuildContext struct {
	Manifest          Manifest
	config            *structs.Config
	runner            string
	longrecipe        []*RecipeWorker
	layers            map[string]*layer.Layer
	touchedAppliances bool
	integrityData     map[string]string
	ldir              string
	homedir           string
}

func NewBuildContext(manifest Manifest, ldir string, config *structs.Config) *BuildContext {
	home, _ := os.UserCacheDir()
	pc := filepath.Join(home, "particle_cache")
	return &BuildContext{
		Manifest:          manifest,
		config:            config,
		runner:            "thin",
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

func (ctx *BuildContext) hookAddLayer(layer *layer.Layer) {
	ctx.layers[layer.Hash] = layer
}

func (ctx *BuildContext) hookPushRecipe(rw *RecipeWorker) {
	ctx.longrecipe = append(ctx.longrecipe, rw)
}
