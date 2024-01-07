package webserver

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/m41denx/particle/utils/fs"
	"github.com/m41denx/particle/webserver/db"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
)

func StartServer(host string, port uint) error {
	app := fiber.New(fiber.Config{
		ServerHeader:          "Particle Repository",
		ETag:                  false,
		BodyLimit:             1024 * 1024,
		ErrorHandler:          nil,
		DisableKeepalive:      false,
		DisableStartupMessage: true,
		AppName:               "Particle Repository",
	})
	app.Use(logger.New())
	app.Use(basicauth.New(basicauth.Config{
		Users: nil,
		Realm: "Particle Repository",
		Authorizer: func(uname string, passwd string) bool {
			return true
		},
		ContextUsername: "username",
		ContextPassword: "password",
	}))
	app.Get("/", func(c *fiber.Ctx) error {
		return nil
	})
	app.Get("/repo/:author/:name\\@:version/:arch",
		func(c *fiber.Ctx) error {
			return c.SendString(fmt.Sprintf("%+v", c.AllParams()))
		})
	app.Post("/upload/:name\\@:version/:arch", nil)
	app.Get("/layers/:layerid", nil)
	app.Get("/user", apiUser)

	app.Group("/upload/")

	fmt.Println(color.CyanString("Starting Particle Repository on http://%s:%d\n[Ctrl+C to stop]", host, port))
	return app.Listen(fmt.Sprintf("%s:%d", host, port))
}

func InitDB(dbtype string, dsn string) (err error) {
	switch dbtype {
	case "mysql":
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "local":
		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	default:
		return fmt.Errorf("unknown Database type: %s", dbtype)
	}
	if err != nil {
		return
	}
	err = DB.AutoMigrate(&db.User{})
	if err != nil {
		return
	}
	err = DB.AutoMigrate(&db.Particle{})
	if err != nil {
		return
	}
	fmt.Println(color.CyanString("Using Database: %s", strings.ToTitle(dbtype)))
	return
}

func InitFS(fstype string, params map[string]string) (err error) {
	switch fstype {
	case "local":
		FS = fs.NewLocalFS()
	case "s3":
		FS = fs.NewS3FS(params)
		LayerDomain = params["domain"]
	default:
		return fmt.Errorf("unknown Filesystem type: %s", fstype)
	}
	fmt.Println(color.CyanString("Using Filesystem: %s", strings.ToTitle(fstype)))
	return nil
}
