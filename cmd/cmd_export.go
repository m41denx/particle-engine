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

func NewCmdExport() *CmdExport {
	cmd := &CmdExport{
		fs: flag.NewFlagSet("export", flag.ExitOnError),
	}

	return cmd
}

type CmdExport struct {
	dir string
	fs  *flag.FlagSet
}

func (cmd *CmdExport) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdExport) Help() string {
	return `
Usage: ` + color.CyanString(binName+" %s <path>", cmd.Name()) + `

  This command exports built(!) particle at <path>.

  Example: ` + binName + ` export ~/particle

  Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdExport) Init(args []string) (err error) {
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

func (cmd *CmdExport) Run() error {
	manif, err := manifest.NewManifestFromFile(filepath.Join(cmd.dir, "particle.yaml"))
	if err != nil {
		return err
	}
	ctx := builder.NewBuildContext(manif, cmd.dir, pkg.Config)
	if err := ctx.Export(); err != nil {
		return err
	}
	return nil
}
