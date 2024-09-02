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

func NewCmdPackage() *CmdPackage {
	cmd := &CmdPackage{
		fs: flag.NewFlagSet("package", flag.ExitOnError),
	}

	return cmd
}

type CmdPackage struct {
	dir string
	fs  *flag.FlagSet
}

func (cmd *CmdPackage) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdPackage) Help() string {
	return `
Usage: ` + color.CyanString(binName+" %s <path>", cmd.Name()) + `

  This command packages built(!) particle at <path>.

  Example: ` + binName + ` package ~/particle

  Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdPackage) Init(args []string) (err error) {
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

func (cmd *CmdPackage) Run() error {
	manif, err := manifest.NewManifestFromFile(filepath.Join(cmd.dir, "particle.yaml"))
	if err != nil {
		return err
	}
	ctx := builder.NewBuildContext(manif, cmd.dir, pkg.Config)
	if err := ctx.Build(); err != nil {
		return err
	}
	return nil
}
