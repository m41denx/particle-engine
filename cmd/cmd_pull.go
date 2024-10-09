package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/m41denx/particle-engine/pkg"
	"github.com/m41denx/particle-engine/pkg/builder"
	"github.com/m41denx/particle-engine/pkg/manifest"
	"github.com/m41denx/particle-engine/utils"
	"os"
	"strings"
)

func NewCmdPull() *CmdPull {
	cmd := &CmdPull{
		fs: flag.NewFlagSet("pull", flag.ExitOnError),
	}
	cmd.fs.StringVar(&cmd.arch, "a", "", fmt.Sprintf("Override architecture. Supported: %s",
		strings.Join(utils.SUPPORTED_ARCH, ", ")))
	return cmd
}

type CmdPull struct {
	arch string
	dir  string
	fs   *flag.FlagSet
}

func (cmd *CmdPull) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdPull) Help() string {
	return `
Usage: ` + color.CyanString(binName+" pull <author/particle@version>") + `

  This command pulls particle layers and all dependencies, but doesn't build them

  Example: ` + binName + ` pull core/builder

  Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdPull) Init(args []string) (err error) {
	err = cmd.fs.Parse(args)
	if err != nil {
		return
	}
	if cmd.fs.NArg() > 0 {
		cmd.dir = cmd.fs.Arg(0)
	} else {
		return fmt.Errorf("particle name is required")
	}
	if len(cmd.arch) > 0 {
		return os.Setenv("PARTICLE_ARCH", cmd.arch)
	}
	return
}

func (cmd *CmdPull) Run() error {
	manif := manifest.NewManifest("particle")
	manif.Recipe = []manifest.RecipeLayerStanza{
		{UseParticle: cmd.dir},
	}
	manif.Meta = manifest.MetaStanza{
		"pull": uuid.NewString(),
	}
	ctx := builder.NewBuildContext(manif, cmd.dir, pkg.Config)
	defer func() { _ = ctx.Clean(false) }()
	if err := ctx.FetchDependencies(); err != nil {
		return err
	}
	if err := ctx.PullOnly(); err != nil {
		return err
	}
	return nil
}
