package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/particle"
	"github.com/m41denx/particle/utils"
	"path"
	"strings"
)

func NewCmdPublish() *CmdPublish {
	cmd := &CmdPublish{
		fs: flag.NewFlagSet("publish", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.name, "n", "", "Override particle name")
	cmd.fs.StringVar(&cmd.version, "v", "", "Override particle version")
	cmd.fs.StringVar(&cmd.url, "u", "", "Override repository URL")

	cmd.fs.BoolVar(&cmd.private, "private", false, "Publish as private")
	cmd.fs.BoolVar(&cmd.unlisted, "unlisted", false, "Publish as unlisted")

	cmd.fs.StringVar(&cmd.arch, "a", "", fmt.Sprintf("Override architecture. Supported: %s",
		strings.Join(utils.SUPPORTED_ARCH, ", ")))

	return cmd
}

type CmdPublish struct {
	fs       *flag.FlagSet
	private  bool
	unlisted bool
	name     string
	version  string
	url      string
	dir      string
	arch     string
}

func (cmd *CmdPublish) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdPublish) Help() string {
	return `
Usage: ` + color.CyanString(binname+" publish <path> [args]") + `

	This command publishes built particle at <path> to repository.

	Example: ` + binname + ` publish ~/particle -n particular -v 1.0 -u https://particles.fruitspace.one

	Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdPublish) Init(args []string) (err error) {
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

func (cmd *CmdPublish) Run() error {
	p, err := particle.NewParticleFromFile(path.Join(cmd.dir, "particle.json"))
	if err != nil {
		return err
	}
	repo := particle.NewRepoMgr().WithPrivate(cmd.private).WithUnlisted(cmd.unlisted).
		WithName(cmd.name).WithVersion(cmd.version).WithUrl(cmd.url).WithArch(cmd.arch)

	err = repo.Publish(p)
	return err
}
