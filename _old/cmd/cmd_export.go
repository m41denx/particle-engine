package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
)

func NewCmdExport() *CmdExport {
	cmd := &CmdExport{
		fs: flag.NewFlagSet("export", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.dest, "dest", ".", "Export directory")

	return cmd
}

type CmdExport struct {
	fs   *flag.FlagSet
	dir  string
	dest string
}

func (cmd *CmdExport) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdExport) Help() string {
	return `
Usage: ` + color.CyanString(binname+" export <path> [args]") + `

	This command exports built particle at <path> to [-dest <dir>].

	Example: ` + binname + ` export ~/particle -dest ./exported

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
	return nil
}
