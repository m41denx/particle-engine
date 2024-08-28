package webserver

import (
	"github.com/gofiber/fiber/v2"
	"github.com/m41denx/particle-engine/pkg/webserver/db"
	"log"
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
	var sz *uint
	// SELECT sum(particle_layers.size) FROM particle_layers JOIN particles ON particle_layers.particle_id=particles.id WHERE particles.uid=1
	if err := DB.Model(db.ParticleLayer{}).Joins("JOIN particles ON particle_layers.particle_id=particles.id").
		Where(db.Particle{UID: user.ID}).Select("sum(particle_layers.size)").Scan(&sz).Error; err != nil {
		log.Println(err)
	}
	if sz == nil {
		sz = new(uint)
	}
	return c.JSON(UserResponse{
		User:     user,
		UsedSize: *sz,
	})
}
