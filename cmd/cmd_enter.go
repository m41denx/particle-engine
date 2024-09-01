package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg"
	"github.com/m41denx/particle-engine/pkg/builder"
	"github.com/m41denx/particle-engine/pkg/manifest"
	"path/filepath"
)

func NewCmdEnter() *CmdEnter {
	cmd := &CmdEnter{
		fs: flag.NewFlagSet("enter", flag.ExitOnError),
	}

	return cmd
}

type CmdEnter struct {
	dir string
	fs  *flag.FlagSet
}

func (cmd *CmdEnter) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdEnter) Help() string {
	return `
Usage: ` + color.CyanString(binName+" enter <path>") + `

  This command enters prepared(!) particle environment at <path>.

  Example: ` + binName + ` enter ~/particle

  Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdEnter) Init(args []string) (err error) {
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

func (cmd *CmdEnter) Run() error {
	manif, err := manifest.NewManifestFromFile(filepath.Join(cmd.dir, "particle.yaml"))
	if err != nil {
		return err
	}
	ctx := builder.NewBuildContext(manif, cmd.dir, pkg.Config)
	if err := ctx.Enter(); err != nil {
		return err
	}
	return nil
}
