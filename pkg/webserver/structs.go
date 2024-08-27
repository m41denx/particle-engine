package webserver

import "github.com/m41denx/particle-engine/pkg/webserver/db"

type UserResponse struct {
	db.User
	UsedSize uint `json:"used_size"`
}

type ErrorResponse struct {
	Message string
}
