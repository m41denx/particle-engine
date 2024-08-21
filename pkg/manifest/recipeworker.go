package manifest

import (
	"context"
	"errors"
	"fmt"
	"github.com/m41denx/particle-engine/pkg/layer"
	"github.com/m41denx/particle-engine/utils"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
)

type RecipeWorker struct {
	ctx      *BuildContext
	parent   *RecipeWorker
	manifest Manifest
	layer    *layer.Layer

	ldir string
}

func NewRecipeWorker(ctx *BuildContext, parent *RecipeWorker, manifest Manifest) *RecipeWorker {
	return &RecipeWorker{
		ctx:      ctx,
		parent:   parent,
		manifest: manifest,
		ldir:     ctx.homedir,
	}
}

func NewRecipeWorkerFromURL(ctx *BuildContext, parent *RecipeWorker, meta ParticleMeta) (*RecipeWorker, error) {
	manifestURL := fmt.Sprintf("%s%s/%s.yaml", meta.Server, meta.Fullname, utils.GetArchString())
	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return nil, err
	}
	req, err = ctx.config.GenerateRequestURL(req, manifestURL)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(context.Background()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	manif, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	err = yaml.Unmarshal(manif, &manifest)
	if err != nil {
		return nil, err
	}
	return NewRecipeWorker(ctx, parent, manifest), nil
}

func (rw *RecipeWorker) fetchChildren() error {
	for _, child := range rw.manifest.Recipe {
		if child.GetParticle() == "blank" {
			continue
		}
		meta, err := ParseParticleURL(child.GetParticle())
		if err != nil {
			return errors.New("invalid particle definition: " + child.GetParticle())
		}
		if rw.parent != nil && child.ApplyParticle != "" {
			break
			// We are getting deep in the tree, ignoring runnables
			// There are only usables and result layer
		}

		worker, err := NewRecipeWorkerFromURL(rw.ctx, rw, meta)
		if err != nil {
			return err
		}
		if err = worker.fetchChildren(); err != nil {
			return err
		}
	}
	// Check valid layers
	if rw.parent != nil && rw.parent.parent != nil {
		// We are 100% not a level-1 layer, ignoring runnables
		if len(rw.manifest.Runnable.Build) > 0 {
			return nil
		}
	}
	if rw.parent != nil {
		rw.layer = layer.NewLayer(rw.manifest.Layer.Block, rw.ldir, rw.manifest.Layer.Server)
		rw.ctx.hookAddLayer(rw.layer)
		rw.ctx.hookPushRecipe(rw)
	}
	if rw.manifest.Runnable.Runner == "full" {
		rw.ctx.runnerType = "full"
	}
	return nil
}
