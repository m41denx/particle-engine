package webserver

import (
	"github.com/m41denx/particle-engine/utils/fs"
	"gorm.io/gorm"
)

var (
	SUPPORTED_ARCH = []string{"w64", "l64", "l64a", "d64", "d64a"}
	DB             *gorm.DB
	FS             fs.FS
	LayerDomain    string
)
