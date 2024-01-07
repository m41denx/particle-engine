package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/particle"
)

func NewCmdPrepare() *CmdPrepare {
	cmd := &CmdPrepare{
		fs: flag.NewFlagSet("prepare", flag.ExitOnError),
	}
	return cmd
}

type CmdPrepare struct {
	dir string
	fs  *flag.FlagSet
}

func (cmd *CmdPrepare) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdPrepare) Help() string {
	return `
Usage: ` + color.CyanString(binname+" prepare <path>") + `

  This command pulls dependencies and prepares particle located at <path>.

  Example: ` + binname + ` prepare ~/particle

  Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdPrepare) Init(args []string) (err error) {
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

func (cmd *CmdPrepare) Run() error {
	particle.ParticlePrepare(cmd.dir)
	return nil
}
