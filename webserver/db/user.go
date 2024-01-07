package db

import gorm "github.com/cradio/gormx"

type User struct {
	gorm.Model
	Username       string
	Token          string
	MaxAllowedSize int64 //bytes
	IsAdmin        bool
}

func (u *User) TableName() string {
	return "particle_users"
}
