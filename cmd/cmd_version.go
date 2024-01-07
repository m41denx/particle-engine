package main

import (
	"fmt"
	"github.com/fatih/color"
)

func NewCmdVersion() *CmdVersion {
	return &CmdVersion{}
}

type CmdVersion struct {
}

func (cmd *CmdVersion) Name() string {
	return "version"
}

func (cmd *CmdVersion) Help() string {
	return ""
}

func (cmd *CmdVersion) Init(args []string) (err error) {
	return
}

func (cmd *CmdVersion) Run() error {
	fmt.Println(color.CyanString("✨ Particle Engine v"+v), "(c) M41den")
	fmt.Println(color.CyanString("Build Tag:\t"), BuildTag)
	fmt.Println(color.CyanString("Build Date:\t"), BuildDate)
	return nil
}
