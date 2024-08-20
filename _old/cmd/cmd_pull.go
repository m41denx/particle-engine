package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/particle"
	"github.com/m41denx/particle/utils"
	"strings"
)

func NewCmdPull() *CmdPull {
	cmd := &CmdPull{
		fs: flag.NewFlagSet("pull", flag.ExitOnError),
	}
	cmd.fs.StringVar(&cmd.url, "u", "", "Override repository URL")

	cmd.fs.StringVar(&cmd.arch, "a", "", fmt.Sprintf("Override architecture. Supported: %s",
		strings.Join(utils.SUPPORTED_ARCH, ", ")))

	return cmd
}

type CmdPull struct {
	fs   *flag.FlagSet
	url  string
	name string
	arch string
}

func (cmd *CmdPull) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdPull) Help() string {
	return `
Usage: ` + color.CyanString(binname+" pull <particle> [args]") + `

	This command pulls particle and its dependencies to local cache.

	Example: ` + binname + ` pull -u https://particles.fruitspace.one -a w32 m41den/busybox@latest

	Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdPull) Init(args []string) (err error) {
	err = cmd.fs.Parse(args)
	if err != nil {
		return
	}
	if cmd.fs.NArg() > 0 {
		cmd.name = cmd.fs.Arg(0)
	} else {
		return fmt.Errorf("particle name is required")
	}
	return
}

func (cmd *CmdPull) Run() (err error) {
	return particle.NewRepoMgr().WithUrl(cmd.url).WithArch(cmd.arch).Pull(cmd.name)
}
