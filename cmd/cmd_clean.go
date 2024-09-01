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

func NewCmdClean() *CmdClean {
	cmd := &CmdClean{
		fs: flag.NewFlagSet("clean", flag.ExitOnError),
	}
	cmd.fs.BoolVar(&cmd.all, "a", false, "Clean all")
	return cmd
}

type CmdClean struct {
	dir string
	all bool
	fs  *flag.FlagSet
}

func (cmd *CmdClean) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdClean) Help() string {
	return `
Usage: ` + color.CyanString(binName+" clean <path>") + `

  This command cleans particle build cache at <path>.

  Example: ` + binName + ` clean ~/particle

  Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdClean) Init(args []string) (err error) {
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

func (cmd *CmdClean) Run() error {
	manif, err := manifest.NewManifestFromFile(filepath.Join(cmd.dir, "particle.yaml"))
	if err != nil {
		return err
	}
	ctx := builder.NewBuildContext(manif, cmd.dir, pkg.Config)
	if err := ctx.Clean(cmd.all); err != nil {
		return err
	}
	return nil
}
