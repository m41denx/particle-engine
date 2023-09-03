package particle

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Engine struct {
	Particle  *Particle
	Runnables map[string]string
	dir       string
}

func NewEngine(particle *Particle) *Engine {
	return &Engine{
		Particle: particle,
		dir:      filepath.Join(particle.dir, "engines", particle.Manifest.Name),
	}
}

func (e *Engine) Load() error {
	data, err := os.ReadFile(filepath.Join(e.dir, "run.json"))
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &e.Runnables)
	if err != nil {
		return err
	}
	for k, v := range e.Runnables {
		p, errx := filepath.Abs(filepath.Join(e.dir, v))
		if errx != nil {
			continue
		}
		e.Runnables[k] = p
	}
	return nil
}
