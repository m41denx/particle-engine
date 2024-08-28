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
	err = DB.Where(db.User{Username: uname, Token: passwd}).First(&user).Error
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
	// SELECT sum(particle_layers.size) FROM particle_layers JOIN particle_repo ON particle_layers.particle_id=particle_repo.id WHERE particle_repo.uid=1
	if err := DB.Model(db.ParticleLayer{}).Joins("JOIN particle_repo ON particle_layers.particle_id=particle_repo.id").
		Where("particle_repo.uid=?", user.ID).Select("sum(particle_layers.size)").Scan(&sz).Error; err != nil {
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
