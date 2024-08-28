package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg"
	"github.com/m41denx/particle-engine/structs"
	"github.com/m41denx/particle-engine/utils"
	"os"
	"path/filepath"
	"runtime/debug"
)

const Version = "0.9.0-dev"

var BuildTag string
var BuildDate string

var binName = filepath.Base(os.Args[0])
var homeDir = utils.PrepareStorage()

func init() {
	var err error
	pkg.Config, err = structs.LoadConfig(filepath.Join(homeDir, "config.json"))
	if err != nil {
		pkg.Config.SaveTo(filepath.Join(homeDir, "config.json"))
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(string(debug.Stack()))
			fmt.Println(err)
			os.Exit(-1)
		}
	}()

	subcommands := []Command{
		NewCmdInit(),
		NewCmdPrepare(),
		NewCmdBuild(),
		NewCmdEnter(),
		NewCmdServe(),
		NewCmdPublish(),
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
				fmt.Println(color.RedString("\nERROR:"), err)
			}
			return
		}
	}
	help()
}

func help() {
	fmt.Print(`
Usage: ` + binName + ` <command> [args]

Commands:
	init		Initializes blank particle directory
	prepare		Prepares particle from manifest
	build		Builds particle after preparations and modifications
	enter		Enters particle environment shell
	export		Exports particle distribution
	auth		Logs you into remote repository account
	publish		Publishes your particle to remote repository
	serve		Starts local repository webserver (see help)

Other commands:
	pull 		Pulls particle and it's dependencies from remote repository without building it
	local		Manages local particles
	version		Prints version
`)
}
