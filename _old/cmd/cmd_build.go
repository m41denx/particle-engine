package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/particle"
)

func NewCmdBuild() *CmdBuild {
	cmd := &CmdBuild{
		fs: flag.NewFlagSet("build", flag.ExitOnError),
	}
	cmd.fs.BoolVar(&particle.UseTerminal, "q", false, "Suppress terminal animations")

	return cmd
}

type CmdBuild struct {
	dir string
	fs  *flag.FlagSet
}

func (cmd *CmdBuild) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdBuild) Help() string {
	return `
Usage: ` + color.CyanString(binname+" build <path>") + `

  This command builds prepared(!) particle at <path>.

  Example: ` + binname + ` build ~/particle

  Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdBuild) Init(args []string) (err error) {
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

func (cmd *CmdBuild) Run() error {
	particle.ParticleBuild(cmd.dir)
	return nil
}
