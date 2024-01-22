package db

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username       string `json:"username"`
	Token          string `json:"-"`
	MaxAllowedSize uint   `json:"max_allowed_size" gorm:"default:0"` //bytes
	IsAdmin        bool   `json:"-"`
}

func (u *User) TableName() string {
	return "particle_users"
}
