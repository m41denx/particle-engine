package manifest

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// const DefaultRepo = "http://particles.fruitspace.one/repo/"
const DefaultRepo = "http://127.0.0.1:8000/"

type Manifest struct {
	Name     string              `yaml:"name" json:"name"`
	Meta     MetaStanza          `yaml:"meta,omitempty" json:"meta,omitempty"`
	Layer    LayerStanza         `yaml:"layer" json:"layer"`
	Recipe   []RecipeLayerStanza `yaml:"recipe" json:"recipe"`
	Runnable RunnableStanza      `yaml:"runnable,omitempty" json:"runnable,omitempty"`
}

type MetaStanza map[string]string

type LayerStanza struct {
	Block  string `yaml:"block" json:"block"`
	Server string `yaml:"server,omitempty" json:"server,omitempty"`
}

// region RecipeLayerStanza

type RecipeLayerStanza struct {
	UseParticle   string     `yaml:"use,omitempty" json:"use,omitempty"`
	ApplyParticle string     `yaml:"apply,omitempty" json:"apply,omitempty"`
	Env           MetaStanza `yaml:"env,omitempty" json:"env,omitempty"`
	Command       string     `yaml:"command,omitempty" json:"command,omitempty"`
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
	Runner  string                `yaml:"runner,omitempty" json:"runner,omitempty"`
	Require []RecipeLayerStanza   `yaml:"require,omitempty" json:"require,omitempty"`
	Build   []RunnableBuildStanza `yaml:"build" json:"build"`
	Expose  MetaStanza            `yaml:"expose,omitempty" json:"expose,omitempty"`
}

type RunnableBuildStanza struct {
	Run             string `yaml:"run,omitempty" json:"run,omitempty"`
	CopySource      string `yaml:"copy,omitempty" json:"copy,omitempty"`
	CopyDestination string `yaml:"to,omitempty" json:"to,omitempty"`
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

func (m *Manifest) ToYaml() string {
	d, _ := yaml.Marshal(m)
	return string(d)
}

func (m *Manifest) ToJson() string {
	d, _ := json.Marshal(m)
	return string(d)
}

func NewManifest(name string) Manifest {
	return Manifest{
		Name: name,
		Meta: MetaStanza{
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

func NewManifestFromFile(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return NewManifest(""), err
	}
	var m Manifest
	err = yaml.Unmarshal(data, &m)
	return m, err
}
