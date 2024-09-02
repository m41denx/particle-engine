package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg"
	"github.com/m41denx/particle-engine/pkg/builder"
	"github.com/m41denx/particle-engine/pkg/manifest"
	"github.com/m41denx/particle-engine/utils"
	"os"
	"path/filepath"
	"strings"
)

func NewCmdBuild() *CmdBuild {
	cmd := &CmdBuild{
		fs: flag.NewFlagSet("build", flag.ExitOnError),
	}
	cmd.fs.StringVar(&cmd.arch, "a", "", fmt.Sprintf("Override architecture. Supported: %s",
		strings.Join(utils.SUPPORTED_ARCH, ", ")))
	cmd.fs.BoolVar(&cmd.clean, "clean", false, "Clean after build (only with export)")
	cmd.fs.BoolVar(&cmd.export, "export", false, "Export package")
	return cmd
}

type CmdBuild struct {
	dir    string
	arch   string
	clean  bool
	export bool
	fs     *flag.FlagSet
}

func (cmd *CmdBuild) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdBuild) Help() string {
	return `
Usage: ` + color.CyanString(binName+" build <path>") + `

  This command pulls dependencies and builds particle located at <path>.

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
	if len(cmd.arch) > 0 {
		return os.Setenv("PARTICLE_ARCH", cmd.arch)
	}
	return
}

func (cmd *CmdBuild) Run() error {
	if cmd.clean && !cmd.export {
		return errors.New("--clean can only be used with --export")
	}
	manif, err := manifest.NewManifestFromFile(filepath.Join(cmd.dir, "particle.yaml"))
	if err != nil {
		return err
	}
	ctx := builder.NewBuildContext(manif, cmd.dir, pkg.Config)
	if err := ctx.FetchDependencies(); err != nil {
		ctx.Clean(false)
		return err
	}
	if err := ctx.PrepareEnvironment(); err != nil {
		ctx.Clean(false)
		return err
	}
	if cmd.export {
		if err = ctx.Export(); err != nil {
			return err
		}
	}
	if cmd.clean {
		return ctx.Clean(false)
	}
	return nil
}
