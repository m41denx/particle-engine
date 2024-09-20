package db

import "gorm.io/gorm"

type Particle struct {
	gorm.Model
	Name        string
	Author      string
	UID         uint
	Description string
	Layers      []*ParticleLayer
	IsPrivate   bool `gorm:"default:0"`
	IsUnlisted  bool `gorm:"default:0"`
}

func (p *Particle) TableName() string {
	return "particle_repo"
}
