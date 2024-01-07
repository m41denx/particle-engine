package main

import "fmt"

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

func (cmd *CmdVersion) Run() error {
	fmt.Println("Particle Engine v0.4 (c) M41den")
	return nil
}
