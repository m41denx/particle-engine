package structs

import (
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

const DefaultRepo = "http://particles.fruitspace.one"

type Manifest struct {
	Type     string        `yaml:"type,omitempty"`
	Name     string        `yaml:"name"`
	Meta     Meta          `yaml:"meta,omitempty"`
	Layer    Layer         `yaml:"layer"`
	Recipe   []RecipeLayer `yaml:"recipe"`
	Runnable Runnable      `yaml:"runnable,omitempty"`
}

type Meta map[string]string

type Layer struct {
	Block  string `yaml:"block"`
	Server string `yaml:"server,omitempty"`
}

// region RecipeLayer

type RecipeLayer struct {
	UseParticle    string `yaml:"use,omitempty"`
	ApplyParticle  string `yaml:"apply,omitempty"`
	EngineParticle string `yaml:"engine,omitempty"`
	Env            Meta   `yaml:"env,omitempty"`
	Command        string `yaml:"command,omitempty"`
}

// endregion

// region Runnable

type Runnable struct {
	Runner  string                `yaml:"runner,omitempty"`
	Require []RecipeLayer         `yaml:"require,omitempty"`
	Build   []RunnableBuildStanza `yaml:"build"`
	Expose  Meta                  `yaml:"expose,omitempty"`
}

type RunnableBuildStanza struct {
	Run             string `yaml:"run,omitempty"`
	CopySource      string `yaml:"copy,omitempty"`
	CopyDestination string `yaml:"to,omitempty"`
}

// endregion

func (m *Manifest) GetServer() string {
	blk := strings.Split(m.Name, "/")
	if len(blk) > 2 {
		return "http://" + strings.Join(blk[:len(blk)-2], "/")
	}
	return DefaultRepo
}

func (m *Manifest) GetLayerServer() string {
	if len(m.Layer.Server) == 0 {
		return DefaultRepo
	}
	return m.Layer.Server
}

func (m *Manifest) SaveTo(dest string) {
	d, _ := yaml.Marshal(m)
	os.WriteFile(dest, d, 0755)
}

func NewManifest(name string) Manifest {
	return Manifest{
		Name: name,
		Meta: Meta{
			"author": "username",
			"note":   "Short note",
		},
		Layer: Layer{
			Block: "[md5 autogen]",
		},
	}
}
