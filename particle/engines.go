package particle

import (
	"encoding/json"
	"github.com/alessio/shellescape"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func PrepareExecutor(dir string, command string) *exec.Cmd {
	os.Setenv("PATH", os.Getenv("PATH")+";"+filepath.Join(dir, "bin"))
	c := strings.Fields(command)
	cmd := exec.Command(c[0], c[1:]...)
	env := os.Environ()
	for k, v := range MetaCache {
		env = append(env, k+"="+shellescape.Quote(v))
	}
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
