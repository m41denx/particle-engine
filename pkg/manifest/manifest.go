package manifest

import (
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// const DefaultRepo = "http://particles.fruitspace.one/repo/"
const DefaultRepo = "http://127.0.0.1:8000/repo/"

type Manifest struct {
	Name     string              `yaml:"name"`
	Meta     Meta                `yaml:"meta,omitempty"`
	Layer    LayerStanza         `yaml:"layer"`
	Recipe   []RecipeLayerStanza `yaml:"recipe"`
	Runnable RunnableStanza      `yaml:"runnable,omitempty"`
}

type Meta map[string]string

func (meta *Meta) asMap() map[string]string {
	return *meta
}

type LayerStanza struct {
	Block  string `yaml:"block"`
	Server string `yaml:"server,omitempty"`
}

// region RecipeLayerStanza

type RecipeLayerStanza struct {
	UseParticle   string `yaml:"use,omitempty"`
	ApplyParticle string `yaml:"apply,omitempty"`
	Env           Meta   `yaml:"env,omitempty"`
	Command       string `yaml:"command,omitempty"`
}

func (m *RecipeLayerStanza) GetParticle() string {
	if len(m.UseParticle) > 0 && len(m.ApplyParticle) > 0 {
		panic("Both `use` and `apply` cannot be defined for a recipe layer")
	}
	if len(m.ApplyParticle) > 0 {
		return m.ApplyParticle
	}
	return m.UseParticle
}

// endregion

// region RunnableStanza

type RunnableStanza struct {
	Runner  string                `yaml:"runner,omitempty"`
	Require []RecipeLayerStanza   `yaml:"require,omitempty"`
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
		Layer: LayerStanza{
			Block: "[sha256 autogen]",
		},
		Recipe: []RecipeLayerStanza{
			{
				UseParticle: "blank",
			},
		},
	}
}
