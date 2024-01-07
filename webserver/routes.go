package webserver

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/m41denx/particle/webserver/db"
)

func StartServer() {
	app := fiber.New()
	app.Get("/repo/:author/:name\\@:version/:arch.json",
		func(c *fiber.Ctx) error {
			return c.SendString(fmt.Sprintf("%+v", c.AllParams()))
		})

	app.Group("/upload/")

	app.Listen(":3000")
}

func InitDB() {
	DB.AutoMigrate(&db.User{})
	DB.AutoMigrate(&db.Particle{})
}
