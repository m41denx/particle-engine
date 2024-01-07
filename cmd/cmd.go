package main

import (
	particle2 "github.com/m41denx/particle/particle"
	"log"
	"path"
)

func CmdPrepare(pathname string) {
	particle, err := particle2.NewParticleFromFile(path.Join(pathname, "particle.json"))
	if err != nil {
		log.Fatalln(err)
	}
	particle.Analyze()
}

func CmdBuild(pathname string) {
	particle, err := particle2.NewParticleFromFile(path.Join(pathname, "particle.json"))
	if err != nil {
		log.Fatalln(err)
	}
	particle.Build()
}
