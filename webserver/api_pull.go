package webserver

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/m41denx/particle/structs"
	"github.com/m41denx/particle/webserver/db"
	"log"
	"path"
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
	if ver == "latest" || ver == "" {
		// Fetch latest matching
		err = DB.Where(db.Particle{
			Name:   c.Params("name"),
			Author: c.Params("author"),
			Arch:   c.Params("arch"),
		}).Order("updated_at DESC").First(&particle).Error
	} else {
		// Strict verison
		err = DB.Where(db.Particle{
			Name:    c.Params("name"),
			Author:  c.Params("author"),
			Version: ver,
			Arch:    c.Params("arch"),
		}).First(&particle).Error
	}
	if err != nil {
		return c.Status(404).JSON(ErrorResponse{
			Message: "Particle not found",
		})
	}

	// Unauthorized for private particles
	if particle.IsPrivate && particle.UID != user.ID {
		return c.Status(404).JSON(ErrorResponse{
			Message: "Particle not found",
		})
	}

	var manifest structs.Manifest
	err = json.Unmarshal([]byte(particle.Recipe), &manifest)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	manifest.Name = particle.Name
	manifest.Author = particle.Author

	return c.JSON(manifest)
}

func apiPullLayer(c *fiber.Ctx) error {
	// No auth
	layerPath := path.Join("/particles/", c.Params("layerid"))
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
