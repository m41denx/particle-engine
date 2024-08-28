package webserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/m41denx/particle-engine/pkg/manifest"
	"github.com/m41denx/particle-engine/pkg/webserver/db"
	"github.com/m41denx/particle-engine/utils"
	"golang.org/x/exp/slices"
	"log"
	"regexp"
	"strings"
)

func apiUpload(c *fiber.Ctx) error {
	c.Accepts("multipart/form-data", "application/json")
	s := strings.ToLower(string(c.Request().Header.Peek("Content-Type")))
	switch s {
	case "application/json":
		return apiUploadManifest(c)
	default:
		return apiUploadLayer(c)
	}
}

func apiUploadManifest(c *fiber.Ctx) error {
	user, err := auth(c)
	if err != nil || user.ID == 0 {
		return c.Status(403).JSON(ErrorResponse{
			Message: "Invalid credentials",
		})
	}

	name := c.Params("name")
	version := c.Params("version")
	arch := c.Params("arch")
	author := c.Params("author")

	if name == "" || version == "" || arch == "" {
		return c.Status(400).JSON(ErrorResponse{
			Message: "Invalid parameters",
		})
	}

	if !slices.Contains(SUPPORTED_ARCH, arch) {
		return c.Status(400).JSON(ErrorResponse{
			Message: "Unsupported architecture",
		})
	}

	preg := regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

	if !preg.MatchString(name) {
		return c.Status(400).JSON(ErrorResponse{
			Message: fmt.Sprintf("Invalid name: %s (Required alphanumeric, _, -, .)", name),
		})
	}

	if !preg.MatchString(version) {
		return c.Status(400).JSON(ErrorResponse{
			Message: fmt.Sprintf("Invalid version: %s (Required alphanumeric, _, -, .)", version),
		})
	}

	var manif manifest.Manifest
	err = c.BodyParser(&manif)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	manif.Meta["author"] = author
	manif.Name = fmt.Sprintf("%s/%s@%s", author, name, version)

	mb, err := json.Marshal(manif)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	recipe := strings.ReplaceAll(string(mb), "\n", "\\n")

	if !user.IsAdmin && author != user.Username {
		// Only admins can upload for other users
		return c.Status(403).JSON(ErrorResponse{
			Message: "You aren't allowed to upload for other users",
		})
	}

	particle := db.Particle{
		Name:        name,
		Author:      author,
		UID:         user.ID,
		Description: fmt.Sprintf("# %s/%s\n--\n", author, manif.Name),
		Layers:      make([]db.ParticleLayer, 0),
		IsPrivate:   c.QueryBool("private"),
		IsUnlisted:  c.QueryBool("unlisted"),
	}

	layer := db.ParticleLayer{
		Arch:    arch,
		LayerID: manif.Layer.Block,
		Version: version,
		Recipe:  recipe,
	}

	var oldParticle db.Particle
	ex := DB.Model(db.Particle{}).Where(db.Particle{
		Name:   particle.Name,
		UID:    particle.UID,
		Author: author,
	}).Preload("Layers").Select("id").Find(&oldParticle).Error

	if ex == nil {
		// particle exists
		particle.ID = oldParticle.ID
		particle.IsPrivate = oldParticle.IsPrivate
		particle.IsUnlisted = oldParticle.IsUnlisted
		particle.Layers = oldParticle.Layers
	}
	for _, p := range particle.Layers {
		if p.Version == version && p.Arch == arch {
			// layer exists
			if p.LayerID != manif.Layer.Block {
				// hashes differ, delete old layer
				if err := FS.DeleteFile(p.LayerID); err != nil {
					log.Println(err)
				}
			}

			ex = errors.New("layer exists")
			layer.ID = p.ID
			layer.Downloads = p.Downloads
			p = layer
		}
	}

	if ex == nil {
		// No such layer exists
		particle.Layers = append(particle.Layers, layer)
	}

	err = DB.Save(&particle).Error
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}
	return nil
}

func apiUploadLayer(c *fiber.Ctx) error {
	user, err := auth(c)
	if err != nil || user.ID == 0 {
		return c.Status(403).JSON(ErrorResponse{
			Message: "Invalid credentials",
		})
	}

	name := c.Params("name")
	version := c.Params("version")
	arch := c.Params("arch")
	author := c.Params("author")

	if name == "" || version == "" || arch == "" {
		return c.Status(400).JSON(ErrorResponse{
			Message: "Invalid parameters",
		})
	}

	if !user.IsAdmin && author != user.Username {
		// Only admins can upload for other users
		return c.Status(403).JSON(ErrorResponse{
			Message: "You aren't allowed to upload for other users",
		})
	}

	metrics := utils.NewGoMetrics()
	defer func() {
		metrics.Done()
		fmt.Println(metrics.DumpText())
	}()

	var particle db.Particle
	if err := DB.Model(db.Particle{}).Where(db.Particle{UID: user.ID, Name: name}).Find(&particle).Error; err != nil {
		return c.Status(401).JSON(ErrorResponse{
			Message: "Particle doesn't exist",
		})
	}

	metrics.NewStep("Counting layers sizes")

	var sz *uint
	// SELECT sum(particle_layers.size) FROM particle_layers JOIN particles ON particle_layers.particle_id=particles.id WHERE particles.uid=1
	if err := DB.Model(db.ParticleLayer{}).Joins("JOIN particles ON particle_layers.particle_id=particles.id").
		Where("particles.uid=?", user.ID).Select("sum(particle_layers.size)").Scan(&sz).Error; err != nil {
		log.Println(err)
	}

	if sz == nil {
		sz = new(uint)
	}
	maxSz := user.MaxAllowedSize - *sz // To check if user is allowed to upload such large file

	metrics.NewStep("Parsing Multipart")

	mpfd, err := c.MultipartForm()
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}
	layers := mpfd.File["layer"]
	if len(layers) == 0 {
		return c.Status(400).JSON(ErrorResponse{
			Message: "No layer provided",
		})
	}
	layer := layers[0]

	if uint(layer.Size) > maxSz {
		return c.Status(400).JSON(ErrorResponse{
			Message: fmt.Sprintf("You don't have available space for that: %.2f MB of %.2f MB",
				float64(layer.Size)/1024/1024, float64(maxSz)/1024/1024),
		})
	}

	layerID := strings.ReplaceAll(layer.Filename, ".7z", "")
	ld, err := layer.Open()
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}
	defer ld.Close()

	var particleLayer db.ParticleLayer
	err = DB.Where(db.ParticleLayer{
		ParticleID: particle.ID,
		Arch:       arch,
		Version:    version,
		LayerID:    layerID,
	}).First(&particleLayer).Error
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	shallow := false
	if particleLayer.Size != 0 && particleLayer.Size == uint(layer.Size) {
		shallow = true
	}

	err = DB.Model(&particleLayer).Updates(db.ParticleLayer{Size: uint(layer.Size)}).Error
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	if shallow {
		return nil // Nothing to do
	}

	metrics.NewStep("Streaming to S3")

	err = FS.PutFileStream(layerID, ld)

	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	return nil
}
