package particle

import "github.com/m41denx/particle/utils"

var ParticleCache map[string]*Particle
var LayerCache map[string]*Layer
var EngineCache map[string]*Engine
var MetaCache map[string]string

var UnzipProvider = utils.New7Zip()
