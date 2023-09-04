package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/particle"
	"os"
)

const v = "0.3"

func init() {
	particle.ParticleCache = make(map[string]*particle.Particle)
	particle.LayerCache = make(map[string]*particle.Layer)
	particle.EngineCache = make(map[string]*particle.Engine)
	particle.MetaCache = make(map[string]string)
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

	fmt.Println(color.CyanString("Particle builder v."+v), "(c) M41den")

	switch flag.Arg(0) {
	case "init":
		CmdInit(flag.Arg(1))
	case "prepare":
		CmdPrepare(flag.Arg(1))
	case "build":
		CmdBuild(flag.Arg(1))
	case "auth":
	case "publish":
		break
	default:
		help()
	}
}

func help() {
	fmt.Println(`Usage:
	particle <init/prepare/build/...> <path>
`)
}
