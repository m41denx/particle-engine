package webserver

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/m41denx/particle-engine/pkg/manifest"
	"github.com/m41denx/particle-engine/pkg/webserver/db"
	"gorm.io/gorm"
	"log"
)

func apiFetchManifest(c *fiber.Ctx) (err error) {
	user, err := auth(c)
	if err != nil {
		log.Println(err)
		return c.Status(403).JSON(ErrorResponse{
			Message: "Invalid credentials",
		})
	}

	ver := c.Params("version")

	var particle db.Particle

	err = DB.Where(db.Particle{
		Name:   c.Params("name"),
		Author: c.Params("author"),
	}).First(&particle).Error

	if err != nil {
		return c.Status(404).JSON(ErrorResponse{
			Message: "Particle not found",
		})
	}

	var layer db.ParticleLayer

	if ver == "latest" || ver == "" {
		// Fetch latest matching
		err = DB.Where(db.ParticleLayer{
			ParticleID: particle.ID,
			Arch:       c.Params("arch"),
		}).Order("updated_at DESC").First(&layer).Error
	} else {
		// Strict verison
		err = DB.Where(db.ParticleLayer{
			ParticleID: particle.ID,
			Version:    ver,
			Arch:       c.Params("arch"),
		}).Order("updated_at DESC").First(&particle).Error
	}
	if err != nil {
		return c.Status(404).JSON(ErrorResponse{
			Message: "Particle version not found",
		})
	}

	// Unauthorized for private particles
	if particle.IsPrivate && particle.UID != user.ID {
		return c.Status(404).JSON(ErrorResponse{
			Message: "Particle not found",
		})
	}

	var manif manifest.Manifest
	err = json.Unmarshal([]byte(layer.Recipe), &manif)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	if err := DB.Model(db.ParticleLayer{}).Where(db.ParticleLayer{LayerID: layer.LayerID}).Update(
		"downloads", gorm.Expr("downloads + 1"),
	).Error; err != nil {
		log.Println(err)
	}

	manif.Name = particle.Name
	manif.Meta["author"] = particle.Author

	return c.SendString(manif.ToYaml())
}

func apiPullLayer(c *fiber.Ctx) error {
	// No auth
	layerPath := c.Params("layerid")
	if LayerDomain != "" {
		return c.Redirect(LayerDomain+layerPath, 301)
	}
	f, sz, err := FS.GetFileStream(layerPath)
	if err != nil {
		return c.Status(404).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.SendStream(f, sz)
}
