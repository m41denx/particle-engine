package particle

import (
	"encoding/json"
	"github.com/m41denx/particle/structs"
	"github.com/m41denx/particle/utils"
	"log"
	"os"
	"path"
)

func ParticleInit(pathname string, name string) {
	if pathname == "" {
		pathname = "."
	}
	os.MkdirAll(pathname, 0750)
	os.Chdir(pathname)
	mani := structs.NewManifest()
	mani.Name = name
	manifest, _ := json.MarshalIndent(mani, "", "\t")
	os.WriteFile("particle.json", manifest, 0750)
	os.MkdirAll("out", 0750)
	os.MkdirAll("bin", 0750)
	os.MkdirAll("dist", 0750)
	os.MkdirAll("src", 0750)
	os.MkdirAll("engines", 0750)
	os.WriteFile(".gitignore", []byte(utils.GitIgnore), 0750)
	utils.Log("Init done at", pathname)
}

func ParticlePrepare(pathname string) {
	p, err := NewParticleFromFile(path.Join(pathname, "particle.json"))
	if err != nil {
		log.Fatalln(err)
	}
	p.Analyze(false)
}

func ParticleBuild(pathname string) {
	p, err := NewParticleFromFile(path.Join(pathname, "particle.json"))
	if err != nil {
		log.Fatalln(err)
	}
	p.Build()
}
