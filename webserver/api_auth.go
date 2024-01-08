package webserver

import (
	"github.com/gofiber/fiber/v2"
	"github.com/m41denx/particle/webserver/db"
)

func auth(c *fiber.Ctx) (user db.User, err error) {
	uname := c.Locals("username").(string)
	passwd := c.Locals("password").(string)
	if uname == "anonymous" {
		return user, nil
	}
	// You're either anonymous or logged in, no invalid users
	err = DB.Where(db.User{Username: uname, Token: passwd}).Find(&user).Error
	if err != nil {
		return user, err
	}
	return user, nil
}

func apiUser(c *fiber.Ctx) error {
	user, err := auth(c)
	if user.ID == 0 || err != nil {
		return c.Status(403).JSON(ErrorResponse{
			Message: "Invalid credentials",
		})
	}
	var sz uint
	err = DB.Model(db.Particle{}).Where(db.Particle{UID: user.ID}).Select("sum(size) as sz").Find(&sz).Error
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(UserResponse{
		User:     user,
		UsedSize: sz,
	})
}
