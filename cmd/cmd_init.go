package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/particle"
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
Usage: ` + color.CyanString(binname+" init <path> [args]") + `

  This command initializes blank particle directory at <path>.

  Example: ` + binname + ` init ~/particle -n particular

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
	particle.ParticleInit(cmd.dir, cmd.name)
	return nil
}
