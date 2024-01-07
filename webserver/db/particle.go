package db

import gorm "github.com/cradio/gormx"

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
	Size        uint // bytes
	IsPrivate   bool
	IsUnlisted  bool
}

func (p *Particle) TableName() string {
	return "particles"
}
