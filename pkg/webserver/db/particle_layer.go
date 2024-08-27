package db

import "gorm.io/gorm"

type ParticleLayer struct {
	gorm.Model
	ParticleID uint
	Arch       string
	LayerID    string
	Version    string
	Recipe     string
	Downloads  uint `gorm:"default:0"`
	Size       uint `gorm:"default:0"` // bytes
}
