package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg/manifest"
	"log"
	"os"
)

func NewCmdInit() *CmdInit {
	cmd := &CmdInit{
		fs: flag.NewFlagSet("init", flag.ExitOnError),
	}
	cmd.fs.StringVar(&cmd.name, "n", "my_particle", "Particle name")

	return cmd
}

type CmdInit struct {
	dir  string
	name string
	fs   *flag.FlagSet
}

func (cmd *CmdInit) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdInit) Help() string {
	return `
Usage: ` + color.CyanString(binName+" init <path> [args]") + `

  This command initializes blank particle directory at <path>.

  Example: ` + binName + ` init ~/particle -n particular@1.0

  Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdInit) Init(args []string) (err error) {
	err = cmd.fs.Parse(args)
	if err != nil {
		return
	}
	if cmd.fs.NArg() > 0 {
		cmd.dir = cmd.fs.Arg(0)
	} else {
		return fmt.Errorf("path is required")
	}
	return
}

func (cmd *CmdInit) Run() error {
	if cmd.dir == "" {
		cmd.dir = "."
	}
	if cmd.name == "" {
		cmd.name = "my_particle@latest"
	}
	_ = os.MkdirAll(cmd.dir, 0750)
	_ = os.Chdir(cmd.dir)
	manif := manifest.NewManifest(cmd.name)
	manif.SaveTo("particle.yaml")
	log.Println("Init done at", cmd.dir)
	return nil
}
