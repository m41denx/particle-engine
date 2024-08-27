package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg"
	"github.com/m41denx/particle-engine/pkg/manifest"
	"path"
)

func NewCmdBuild() *CmdBuild {
	cmd := &CmdBuild{
		fs: flag.NewFlagSet("build", flag.ExitOnError),
	}

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
Usage: ` + color.CyanString(binName+" build <path>") + `

  This command builds prepared(!) particle at <path>.

  Example: ` + binName + ` build ~/particle

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
	manif, err := manifest.NewManifestFromFile(path.Join(cmd.dir, "particle.yaml"))
	if err != nil {
		return err
	}
	ctx := builder.NewBuildContext(manif, cmd.dir, pkg.Config)
	if err := ctx.Build(); err != nil {
		return err
	}
	return nil
}
