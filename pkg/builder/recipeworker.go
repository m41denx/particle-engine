package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/m41denx/particle-engine/pkg/layer"
	manifest2 "github.com/m41denx/particle-engine/pkg/manifest"
	"github.com/m41denx/particle-engine/utils"
	"gopkg.in/yaml.v3"
	"io"
	"maps"
	"net/http"
	"path"
	"slices"
)

type RecipeWorker struct {
	ctx      *BuildContext
	parent   *RecipeWorker
	manifest manifest2.Manifest
	layer    *layer.Layer

	homedir string
}

func NewRecipeWorker(ctx *BuildContext, parent *RecipeWorker, manifest manifest2.Manifest) *RecipeWorker {
	return &RecipeWorker{
		ctx:      ctx,
		parent:   parent,
		manifest: manifest,
		homedir:  ctx.homedir,
	}
}

func NewRecipeWorkerFromURL(ctx *BuildContext, parent *RecipeWorker, meta manifest2.ParticleMeta) (*RecipeWorker, error) {
	if meta.Fullname == "blank" {
		return NewRecipeWorker(ctx, parent, manifest2.Manifest{
			Name: "blank",
		}), nil
	}

	manifestURL := fmt.Sprintf("%s/repo/%s/%s.yaml", meta.Server, meta.Fullname, utils.GetArchString())
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
	var manifest manifest2.Manifest
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
		meta, err := manifest2.ParseParticleURL(child.GetParticle())
		if err != nil {
			return errors.New("invalid particle definition: " + child.GetParticle())
		}
		if rw.parent != nil && child.ApplyParticle != "" {
			break
			// We are getting deep in the tree, ignoring runnables
			// There are only usables and result layer
		}
		// TODO: Acccount for Appliances' dependencies and somehow shove them in

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
		rw.layer = layer.NewLayer(rw.manifest.Layer.Block, rw.homedir, rw.manifest.Layer.Server)
		rw.ctx.hookAddLayer(rw.layer)
		rw.ctx.hookPushRecipe(rw)
	}
	if rw.manifest.Runnable.Runner == "full" {
		rw.ctx.runnerType = "full"
	}
	return nil
}

func (rw *RecipeWorker) ExtractLayer(destdir string, isRootfs bool) error {
	if rw.manifest.Name == "blank" {
		return nil
	}
	if isRootfs {
		return rw.layer.ExtractTo(destdir)
	}
	if len(rw.manifest.Runnable.Build) > 0 {
		// When pulling from origin, Manifest.Name is replaced by full particle name with full_url
		return rw.layer.ExtractTo(path.Join(destdir, "runnable", rw.manifest.Name)) //FIXME
		// TODO: Allow appliances to expose their commands
	} else {
		return rw.layer.ExtractTo(path.Join(destdir, "build"))
	}
}

func (rw *RecipeWorker) RunAppliance(destdir string, override manifest2.RecipeLayerStanza) error {
	if len(rw.manifest.Runnable.Build) == 0 {
		return nil // Might as well be static particle
	}

	globEnv := map[string]string{}
	if override.Env != nil {
		globEnv = maps.Clone(override.Env)
	}
	if rw.ctx.runnerType == "full" {
		maps.Copy(globEnv, map[string]string{
			"ROOTFS": path.Join("/"),
			"BUILD":  path.Join("/", "build"),
			"MOD":    path.Join("/", "runnable", rw.manifest.Name),
		})
	} else {
		// Busybox maps root as Windows fullpath
		maps.Copy(globEnv, map[string]string{
			"ROOTFS": path.Join(destdir),
			"BUILD":  path.Join(destdir, "build"),
			"MOD":    path.Join(destdir, "runnable", rw.manifest.Name),
		})
	}

	actions := slices.Clone(rw.manifest.Runnable.Build)

	// Weird override af, but who knows
	if override.Command != "" {
		actions = []manifest2.RunnableBuildStanza{
			{
				Run: override.Command,
			},
		}
	}

	for _, action := range actions {
		// Set env for copying
		env := maps.Clone(globEnv)

		if len(action.Run) > 0 {
			// If we run, we run
			if err := rw.ctx.runnerInstance.Run(action.Run, env); err != nil {
				return err
			}
			continue
		}

		// Else we copy files
		if len(action.CopySource) == 0 || len(action.CopyDestination) == 0 {
			return errors.New("action copy source or destination are not specified")
		}
		if err := rw.ctx.runnerInstance.Copy(action.CopySource, action.CopyDestination, env); err != nil {
			return err
		}
	}

	return nil
}
