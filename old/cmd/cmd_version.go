package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/utils"
)

func NewCmdVersion() *CmdVersion {
	cmd := &CmdVersion{
		fs: flag.NewFlagSet("version", flag.ExitOnError),
	}
	cmd.fs.BoolVar(&cmd.update, "update", false, "Update to latest version")
	return cmd
}

type CmdVersion struct {
	fs     *flag.FlagSet
	update bool
}

func (cmd *CmdVersion) Name() string {
	return "version"
}

func (cmd *CmdVersion) Help() string {
	return ""
}

func (cmd *CmdVersion) Init(args []string) (err error) {
	return cmd.fs.Parse(args)
}

func (cmd *CmdVersion) Run() error {
	fmt.Println(color.CyanString("✨ Particle Engine v"+v), "(c) M41den")
	fmt.Println(color.CyanString("Build Tag:\t"), BuildTag)
	fmt.Println(color.CyanString("Build Date:\t"), BuildDate)
	if cmd.update {
		return utils.SelfUpdate(v)
	}
	return nil
}
