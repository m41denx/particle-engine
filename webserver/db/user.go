package db

import gorm "github.com/cradio/gormx"

type User struct {
	gorm.Model
	Username       string `json:"username"`
	Token          string `json:"-"`
	MaxAllowedSize uint   `json:"max_allowed_size"` //bytes
	IsAdmin        bool   `json:"-"`
}

func (u *User) TableName() string {
	return "particle_users"
}
