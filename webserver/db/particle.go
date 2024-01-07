package db

import gorm "github.com/cradio/gormx"

type Particle struct {
	gorm.Model
	Name        string
	Author      string
	UID         int64
	Arch        string
	Version     string
	Description string
	Recipe      string
	Size        int64 // bytes
	IsPrivate   bool
}

func (p *Particle) TableName() string {
	return "particles"
}
