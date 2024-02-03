package particle

import (
	"github.com/m41denx/particle/structs"
	"github.com/m41denx/particle/utils"
	"os"
	"runtime"
	"strconv"
)

var ParticleCache map[string]*Particle
var LayerCache map[string]*Layer
var EngineCache map[string]*Engine
var MetaCache map[string]string

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
	if ncpu > 8 {
		ncpu = 8
	}
	return ncpu
}()
