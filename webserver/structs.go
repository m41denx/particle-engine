package webserver

import "github.com/m41denx/particle/webserver/db"

type UserResponse struct {
	db.User
	UsedSize uint
}

type ErrorResponse struct {
	Message string
}
