package structs

import (
	"encoding/json"
	"os"
)

//const DefaultRepo = "https://raw.githubusercontent.com/m41denx/particle-repo/master/"

const DefaultRepo = "http://127.0.0.1:8000/"

type Manifest struct {
	Name        string            `json:"name"`
	Author      string            `json:"author"`
	Note        string            `json:"note"`
	Block       string            `json:"block"`
	Server      string            `json:"server,omitempty"`
	LayerServer string            `json:"block_server,omitempty"`
	Meta        map[string]string `json:"meta"`
	Recipe      Recipe            `json:"recipe"`
}

type Recipe struct {
	Base    string   `json:"base"`
	Apply   []string `json:"apply"`
	Engines []string `json:"engines"`
	Run     []string `json:"run"`
}

func NewManifest() Manifest {
	return Manifest{
		Name:   "my_particle@1.0",
		Author: "user",
		Note:   "Short description or note about usage",
		Block:  "[md5 autogen]",
		Server: "",
		Meta:   map[string]string{},
		Recipe: Recipe{
			Base:    "blank",
			Apply:   []string{},
			Engines: []string{},
			Run:     []string{},
		},
	}
}

func (m *Manifest) GetServer() string {
	if len(m.Server) == 0 {
		return DefaultRepo
	}
	return m.Server
}

func (m *Manifest) GetLayerServer() string {
	if len(m.LayerServer) == 0 {
		return DefaultRepo
	}
	return m.LayerServer
}

func (m *Manifest) SaveTo(dest string) {
	d, _ := json.MarshalIndent(m, "", "\t")
	os.WriteFile(dest, d, 0755)
}
