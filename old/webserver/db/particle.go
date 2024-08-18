package db

import "gorm.io/gorm"

type Particle struct {
	gorm.Model
	Name        string
	Author      string
	UID         uint
	Arch        string
	LayerID     string
	Version     string
	Description string
	Recipe      string
	Downloads   uint `gorm:"default:0"`
	Size        uint `gorm:"default:0"` // bytes
	IsPrivate   bool `gorm:"default:0"`
	IsUnlisted  bool `gorm:"default:0"`
}

func (p *Particle) TableName() string {
	return "particles"
}