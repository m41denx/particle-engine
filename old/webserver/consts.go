package webserver

import (
	"github.com/m41denx/particle/utils/fs"
	"gorm.io/gorm"
)

var (
	SUPPORTED_ARCH = []string{"w32", "w64", "l64", "l64a", "l32", "d64", "d64a"}
	DB             *gorm.DB
	FS             fs.FS
	LayerDomain    string
)
