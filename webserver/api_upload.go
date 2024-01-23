package webserver

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/m41denx/particle/structs"
	"github.com/m41denx/particle/utils"
	"github.com/m41denx/particle/webserver/db"
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
	if user.ID == 0 || err != nil {
		return c.Status(403).JSON(ErrorResponse{
			Message: "Invalid credentials",
		})
	}

	name := c.Params("name")
	version := c.Params("version")
	arch := c.Params("arch")

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
			Message: "Invalid name (Required alphanumeric, _, -, .)",
		})
	}

	if !preg.MatchString(version) {
		return c.Status(400).JSON(ErrorResponse{
			Message: "Invalid version (Required alphanumeric, _, -, .)",
		})
	}

	var manifest structs.Manifest
	err = c.BodyParser(&manifest)
	if err != nil {
		return c.Status(400).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	manifest.Author = user.Username
	manifest.Name = name + "@" + version

	mb, err := json.Marshal(manifest)
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	recipe := strings.ReplaceAll(string(mb), "\n", "\\n")

	particle := db.Particle{
		Name:        name,
		Author:      user.Username,
		UID:         user.ID,
		Arch:        arch,
		LayerID:     manifest.Block,
		Version:     version,
		Description: manifest.Note,
		Recipe:      recipe,
		Size:        0,
		IsPrivate:   c.QueryBool("private"),
		IsUnlisted:  c.QueryBool("unlisted"),
	}

	var oldParticle db.Particle
	ex := DB.Model(db.Particle{}).Where(db.Particle{
		Name:    particle.Name,
		UID:     particle.UID,
		Version: particle.Version,
		Arch:    arch,
	}).Select("id").Find(&oldParticle).Error

	if ex == nil {
		particle.ID = oldParticle.ID
		//Remove
		err = FS.DeleteFile(oldParticle.LayerID)
		if err != nil {
			log.Println(err)
		}
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
	if user.ID == 0 || err != nil {
		return c.Status(403).JSON(ErrorResponse{
			Message: "Invalid credentials",
		})
	}

	name := c.Params("name")
	version := c.Params("version")
	arch := c.Params("arch")

	if name == "" || version == "" || arch == "" {
		return c.Status(400).JSON(ErrorResponse{
			Message: "Invalid parameters",
		})
	}

	metrics := utils.NewGoMetrics()
	defer func() {
		metrics.Done()
		fmt.Println(metrics.DumpText())
	}()

	var sz *uint
	DB.Model(db.Particle{}).Where(db.Particle{UID: user.ID}).Select("sum(size)").Scan(&sz)

	if sz == nil {
		sz = new(uint)
	}
	maxSz := user.MaxAllowedSize - *sz // To check if user is allowed to upload such large file

	metrics.NewStep("Parsing Multipart")

	mpfd, err := c.MultipartForm()
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

	layerID := layer.Filename
	ld, err := layer.Open()
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	var particle db.Particle
	err = DB.Where(db.Particle{
		Name:    name,
		Author:  user.Username,
		Version: version,
		Arch:    arch,
		LayerID: layerID,
	}).Find(&particle).Error
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
	}

	particle.Size = uint(layer.Size)

	err = DB.Updates(particle).Error
	if err != nil {
		return c.Status(500).JSON(ErrorResponse{
			Message: err.Error(),
		})
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
