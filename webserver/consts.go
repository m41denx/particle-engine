package webserver

import gorm "github.com/cradio/gormx"

var (
	SUPPORTED_ARCH = []string{"w32", "w64", "l64", "l64a", "l32", "d64", "d64a"}
)

var DB *gorm.DB
