package main

import (
	"flag"
	"fmt"
	"github.com/m41denx/particle/particle"
	"github.com/m41denx/particle/utils"
	"os"
)

func init() {
	particle.ParticleCache = make(map[string]*particle.Particle)
	particle.LayerCache = make(map[string]*particle.Layer)
	particle.EngineCache = make(map[string]*particle.Engine)
}

func main() {
	defer func() {
		if recover() != nil {
			fmt.Println()
			os.Exit(0)
		}
	}()
	flag.Parse()

	PrepareStorage()

	switch flag.Arg(0) {
	case "init":
		CmdInit(flag.Arg(1))
	case "prepare":
		CmdPrepare(flag.Arg(1))
	case "build":
	case "auth":
	case "publish":
		break
	default:
		utils.Log(flag.Arg(0))
	}
}
