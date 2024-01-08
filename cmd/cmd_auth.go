package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/particle"
	"github.com/m41denx/particle/structs"
	"io"
	"net/http"
	"path/filepath"
)

func NewCmdAuth() *CmdAuth {
	cmd := &CmdAuth{
		fs: flag.NewFlagSet("auth", flag.ExitOnError),
	}

	cmd.fs.StringVar(&cmd.url, "u", structs.DefaultRepo, "Repository URL")
	cmd.fs.StringVar(&cmd.uname, "a", "", "Username")
	cmd.fs.StringVar(&cmd.token, "t", "", "Token")

	return cmd
}

type CmdAuth struct {
	fs    *flag.FlagSet
	url   string
	uname string
	token string
}

func (cmd *CmdAuth) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdAuth) Help() string {
	return `
Usage: ` + color.CyanString(binname+" auth <url> [args]") + `

	This command authenticates with <url> repository server.

	Authentication example: ` + binname + ` auth -a user -t token

	User info example: ` + binname + ` auth -u https://particles.fruitspace.one

	Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdAuth) Init(args []string) (err error) {
	err = cmd.fs.Parse(args)
	return
}

func (cmd *CmdAuth) Run() error {
	if len(cmd.uname) > 0 || len(cmd.token) > 0 {
		particle.Config.AddRepo(cmd.url, structs.RepoConfig{
			Username: cmd.uname,
			Token:    cmd.token,
		})
	}
	r, _ := http.NewRequest("GET", cmd.url+"/user", nil)
	creds, err := particle.Config.GenerateRequestURL(r, cmd.url)
	if err != nil {
		return err
	}
	data, err := http.DefaultClient.Do(creds)
	if err != nil {
		return err
	}
	defer data.Body.Close()
	b, err := io.ReadAll(data.Body)
	if err != nil {
		return err
	}
	if data.StatusCode != 200 {
		return fmt.Errorf("Auth Failed:\n%s", string(b))
	}
	var user structs.UserResponse
	_ = json.Unmarshal(b, &user)
	fmt.Println(color.CyanString("Authenticated as: "), user.Username)
	fmt.Println(color.CyanString("Used space at %s:", cmd.url),
		fmt.Sprintf("%.2fMB/%.2fMB", float64(user.UsedSize)/1024/1024, float64(user.MaxAllowedSize)/1024/1024))

	particle.Config.SaveTo(filepath.Join(ldir, "config.json"))
	return nil
}
