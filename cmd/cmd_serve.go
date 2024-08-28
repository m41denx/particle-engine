package main

import (
	"flag"
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg/webserver"
)

func NewCmdServe() *CmdServe {
	cmd := &CmdServe{
		fs: flag.NewFlagSet("serve", flag.ExitOnError),
	}
	cmd.fs.StringVar(&cmd.host, "h", "127.0.0.1", "Host")
	cmd.fs.UintVar(&cmd.port, "p", 3000, "Port")

	cmd.fs.StringVar(&cmd.dbm, "dbm", "local", "DB backend (mysql/local)")
	cmd.fs.StringVar(&cmd.dbDSN, "dbaddr", "local.db", "DB DSN (user:pass@tcp(ADDR:PORT)/dbname")

	cmd.fs.StringVar(&cmd.fsm, "fsm", "local", "File Storage backend (s3/local)")
	cmd.fs.StringVar(&cmd.fs3AccessKey, "fs3-key", "", "S3 access key")
	cmd.fs.StringVar(&cmd.fs3SecretKey, "fs3-secret", "", "S3 secret key")
	cmd.fs.StringVar(&cmd.fs3Region, "fs3-region", "us-east-1", "S3 region")
	cmd.fs.StringVar(&cmd.fs3Bucket, "fs3-bucket", "particles", "S3 bucket")
	cmd.fs.StringVar(&cmd.fs3Endpoint, "fs3-endpoint", "https://s3.amazonaws.com", "S3 endpoint")
	cmd.fs.StringVar(&cmd.fs3Domain, "fs3-domain", "https://hub.fruitspace.one", "S3 layers domain")
	cmd.fs.StringVar(&cmd.fs3Prefix, "fs3-prefix", "/layers", "S3 prefix")

	return cmd
}

type CmdServe struct {
	fs           *flag.FlagSet
	host         string // WS host
	port         uint   // WS port
	dbm          string // DB backend (mysql/local)
	dbDSN        string // DB DSN
	fsm          string // File Storage backend (s3/local)
	fs3AccessKey string // S3 access key
	fs3SecretKey string // S3 secret key
	fs3Region    string // S3 region
	fs3Bucket    string // S3 bucket
	fs3Endpoint  string // S3 endpoint
	fs3Domain    string // S3 domain
	fs3Prefix    string // S3 prefix
}

func (cmd *CmdServe) Name() string {
	return cmd.fs.Name()
}

func (cmd *CmdServe) Help() string {
	return `
Usage: ` + color.CyanString(binName+" serve [args]") + `

  This command starts particle repository server.

  Example for dev environment:
    ` + binName + ` serve -p 80
  Example for scalable production deployment:
	` + binName + ` serve -p 3000 -h 0.0.0.0 -dbm mysql -dbaddr "user:password@tcp(127.0.0.1:3306)/particles" \
		-fsm s3 -fs3-key s3key -fs3-secret s3secret -fs3-region us-east-1 -fs3-bucket particles \
        -fs3-endpoint https://s3.amazonaws.com -fs3-domain https://hub.fruitspace.one

  Please see the individual subcommand help for detailed usage information.
`
}

func (cmd *CmdServe) Init(args []string) (err error) {
	return cmd.fs.Parse(args)
}

func (cmd *CmdServe) Run() (err error) {
	err = webserver.InitDB(cmd.dbm, cmd.dbDSN)
	if err != nil {
		return
	}
	err = webserver.InitFS(cmd.fsm, map[string]string{
		"endpoint":   cmd.fs3Endpoint,
		"access_key": cmd.fs3AccessKey,
		"secret":     cmd.fs3SecretKey,
		"region":     cmd.fs3Region,
		"bucket":     cmd.fs3Bucket,
		"domain":     cmd.fs3Domain,
		"prefix":     cmd.fs3Prefix,
	})
	if err != nil {
		return
	}
	err = webserver.StartServer(cmd.host, cmd.port)
	return
}
