package main

import (
	"fmt"
	"github.com/m41denx/particle/particle"
	"github.com/m41denx/particle/structs"
	"github.com/m41denx/particle/utils"
	"os"
	"path/filepath"
)

const v = "0.5.8-dev"

var binname string
var BuildTag string
var BuildDate string
var ldir string

func init() {
	particle.ParticleCache = make(map[string]*particle.Particle)
	particle.LayerCache = make(map[string]*particle.Layer)
	particle.EngineCache = make(map[string]*particle.Engine)
	particle.MetaCache = make(map[string]string)

	binname = filepath.Base(os.Args[0])

	ldir = utils.PrepareStorage()
}

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}()
	var err error
	particle.Config, err = structs.LoadConfig(filepath.Join(ldir, "config.json"))
	if err != nil {
		particle.Config.SaveTo(filepath.Join(ldir, "config.json"))
	}

	subcommands := []Command{
		NewCmdInit(),
		NewCmdPrepare(),
		NewCmdBuild(),
		NewCmdExport(),
		NewCmdAuth(),
		NewCmdPublish(),
		NewCmdServe(),

		NewCmdVersion(),
	}

	if len(os.Args) < 2 {
		help()
		return
	}

	for _, sub := range subcommands {
		if sub.Name() == os.Args[1] {
			err := sub.Init(os.Args[2:])
			if err != nil {
				fmt.Println("Error:", err, "\n", sub.Help())
				return
			}
			err = sub.Run()
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}
	help()
}

func help() {
	fmt.Print(`
Usage: ` + binname + ` <command> [args]

Commands:
	init		Initializes blank particle directory
	prepare		Prepares particle from manifest
	build		Builds particle after preparations and modifications
	export		Exports particle distribution
	auth		Logs you into remote repository account
	publish		Publishes your particle to remote repository
	serve		Starts local repository webserver (see help)

Other commands:
	[unsupported] pull 		Pulls particle and it's dependencies from remote repository without building it
	[unsupported] local		Manages local particles]
	version		Prints version
`)
}
