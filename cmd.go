package main

import (
	"encoding/json"
	particle2 "github.com/m41denx/particle/particle"
	"github.com/m41denx/particle/structs"
	"github.com/m41denx/particle/utils"
	"log"
	"os"
	"path"
)

const GitIgnore = `bin/
dist/
engines/
out/
src/`

func CmdInit(pathname string) {
	if pathname == "" {
		pathname = "."
	}
	os.MkdirAll(pathname, 0750)
	os.Chdir(pathname)
	manifest, _ := json.MarshalIndent(structs.NewManifest(), "", "\t")
	os.WriteFile("particle.json", manifest, 0750)
	os.MkdirAll("out", 0750)
	os.MkdirAll("bin", 0750)
	os.MkdirAll("dist", 0750)
	os.MkdirAll("src", 0750)
	os.MkdirAll("engines", 0750)
	os.WriteFile(".gitignore", []byte(GitIgnore), 0750)
	utils.Log("Init done at", pathname)
}

func PrepareStorage() string {
	home, _ := os.UserCacheDir()
	pc := path.Join(home, "particle_cache")
	os.MkdirAll(pc, 0750)
	os.MkdirAll(path.Join(pc, "layers"), 0750)
	os.MkdirAll(path.Join(pc, "repo"), 0750)
	return pc
}

func CmdPrepare(pathname string) {
	particle, err := particle2.NewParticleFromFile(path.Join(pathname, "particle.json"))
	if err != nil {
		log.Fatalln(err)
	}
	particle.Analyze()
}
