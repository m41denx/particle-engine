package pkg

import (
	"github.com/m41denx/particle-engine/pkg/cache"
	"github.com/m41denx/particle-engine/structs"
	"github.com/m41denx/particle-engine/utils"
	"os"
	"runtime"
	"strconv"
)

var UnzipProvider = utils.New7Zip()

var Config = structs.NewConfig()

var NUMCPU = func() int {
	numcpu, ok := os.LookupEnv("NUMCPU")
	var ncpu int
	if !ok || numcpu == "" {
		ncpu = runtime.NumCPU()
	} else {
		ncpu, _ = strconv.Atoi(numcpu)
	}
	if ncpu > 4 || ncpu < 1 {
		ncpu = 4
	}
	return ncpu
}()

var SharedCache cache.Cache
