package particle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/structs"
	"github.com/m41denx/particle/utils"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type Particle struct {
	Manifest structs.Manifest

	base          *Particle
	engines       []*Particle
	layers        []*Particle
	integrityData map[string]string //filepath -> md5

	dir string
}

func NewParticleFromString(manifest string) (*Particle, error) {
	p := Particle{
		engines: make([]*Particle, 1),
		layers:  make([]*Particle, 1),
	}
	err := json.Unmarshal([]byte(manifest), &p.Manifest)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func NewParticleFromFile(filename string) (*Particle, error) {
	p, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	particle, err := NewParticleFromString(string(p))
	if err == nil {
		particle.dir = path.Dir(filename)
	}
	return particle, err
}

func (p *Particle) SetDir(dir string) {
	p.dir = dir
}

func (p *Particle) Analyze() {
	progress := utils.NewTreeProgress()
	fmt.Println(color.CyanString("Generating dependency tree..."))
	fmt.Println("~/" + p.Manifest.Name)
	p.makeTree(progress)
	fmt.Println(color.CyanString("Downloading dependencies..."))
	p.downloadLayers()
	if p.base != nil {
		fmt.Println(color.CyanString("Extracting base..."))
		p.base.prepareBase("dist") // We build new particle, so we need to operate on it's base rather than itself
	}
	p.calculateIntegrityHash()

	fmt.Println(color.CyanString("Warming up engines..."))
	p.prepareEngines()
	p.populateEngineBinaries()

	fmt.Println(color.CyanString("Preparing applicables..."))
	p.prepareApplicable()

	fmt.Println(color.CyanString("Executing applicables..."))
	p.executeApplicable()
}

func (p *Particle) Build() {
	fmt.Println(color.CyanString("Building particle..."))
	p.compareIntegrity()
	progress := utils.NewTreeProgress()
	progress.TrackFunction("Building layer...", p.makeLayer)
	p.Manifest.SaveTo(filepath.Join(p.dir, "particle.json"))
}

func (p *Particle) makeLayer() {
	l := NewLayer(p.Manifest.GetLayerServer())
	err := l.CreateLayer(filepath.Join(p.dir, "out"), filepath.Join("_tmp_layer_.7z"))
	if err != nil {
		fmt.Println(color.RedString("\nERROR: "), err)
		panic(err)
	}
	h, err := utils.CalcFileHash(filepath.Join(p.dir, "out", "_tmp_layer_.7z"))
	if err != nil {
		fmt.Println(color.RedString("\nERROR: "), err)
		panic(err)
	}
	err = os.Rename(filepath.Join(p.dir, "out", "_tmp_layer_.7z"), filepath.Join(p.dir, h))
	if err != nil {
		fmt.Println(color.RedString("\nERROR: "), err)
		panic(err)
	}
	p.Manifest.Block = h
}

func (p *Particle) makeTree(progress *utils.TreeProgress) {
	var wg sync.WaitGroup

	progress.Tab()
	// Base
	{
		wg.Add(1)
		c := make(chan bool)
		go progress.Run("└── base: "+p.Manifest.Recipe.Base, &wg, c)
		particle, err := p.loadParticle(p.Manifest.Recipe.Base)
		if err != nil {
			log.Fatalln(err)
		}
		c <- true
		wg.Wait()
		p.base = particle
		if particle != nil {
			particle.makeTree(progress)
		}
	}
	progress.Ret()

	// Engines
	progress.Tab()
	for _, engine := range p.Manifest.Recipe.Engines {
		wg.Add(1)
		c := make(chan bool)
		go progress.Run("└── engine: "+engine, &wg, c)
		particle, err := p.loadParticle(engine)
		if err != nil {
			log.Fatalln(err)
		}
		if len(particle.Manifest.Recipe.Run) > 0 {
			log.Fatalln("Engines cannot contain RUN sections in their manifests")
		}
		c <- true
		wg.Wait()
		if particle != nil {
			p.engines = append(p.engines, particle)
			particle.makeTree(progress)
		}
	}
	progress.Ret()

	// Layers to apply
	progress.Tab()
	for _, layer := range p.Manifest.Recipe.Apply {
		wg.Add(1)
		c := make(chan bool)
		go progress.Run("└── apply: "+layer, &wg, c)
		particle, err := p.loadParticle(layer)
		if err != nil {
			log.Fatalln(err)
		}
		c <- true
		wg.Wait()
		if particle != nil {
			p.layers = append(p.layers, particle)
			particle.makeTree(progress)
		}
	}
	progress.Ret()
}

func (p *Particle) downloadLayers() {
	for k, particle := range ParticleCache {
		fmt.Print(color.GreenString("→ %s [%s]", k, particle.Manifest.Block))
		l := NewLayer(particle.Manifest.GetLayerServer())
		err := l.Fetch(particle.Manifest.Block)
		if err != nil {
			fmt.Println(color.RedString("\nERROR: "), err)
			panic(err)
		}
		LayerCache[l.ID] = l
	}
}

func (p *Particle) prepareBase(target string) {
	if p.base != nil {
		p.base.prepareBase(target)
	}
	progress := utils.NewTreeProgress()

	progress.TrackFunction(fmt.Sprintf("→ %s [%s]...", p.Manifest.Name, p.Manifest.Block), func() {
		layer, ok := LayerCache[p.Manifest.Block]
		if !ok {
			fmt.Println(color.RedString("\nERROR: "), p.Manifest.Block+": layer was not found. Not sure what to do...")
			panic(ok)
		}
		err := layer.ExtractTo(filepath.Join(p.dir, target))
		if err != nil {
			fmt.Println(color.RedString("\nERROR: "), err)
			panic(err)
		}
	})
}

func (p *Particle) prepareEngines() {
	// Active layers' engines first
	for _, act := range p.layers {
		if act == nil {
			continue
		}
		act.prepareEngines()
	}
	// Now the real engines
	for _, eng := range p.engines {
		if eng == nil {
			continue
		}
		// Yeah, those too
		eng.prepareEngines()
		if eng.base != nil {
			eng.base.prepareBase(filepath.Join("engines", eng.Manifest.Name))
		}

		progress := utils.NewTreeProgress()
		progress.TrackFunction(fmt.Sprintf("→ %s [%s]...", eng.Manifest.Name, eng.Manifest.Block), func() {
			layer, ok := LayerCache[eng.Manifest.Block]
			if !ok {
				fmt.Println(color.RedString("\nERROR: "), eng.Manifest.Block+": layer was not found. Not sure what to do...")
				panic(ok)
			}
			err := layer.ExtractTo(filepath.Join(p.dir, "engines", eng.Manifest.Name))
			if err != nil {
				fmt.Println(color.RedString("\nERROR: "), err)
				panic(err)
			}
			engine := NewEngine(eng)
			err = engine.Load()
			if err != nil {
				fmt.Println(color.RedString("\nERROR: "), err)
				panic(err)
			}
			EngineCache[eng.Manifest.Name] = engine
			for k, v := range eng.Manifest.Meta {
				MetaCache[k] = os.ExpandEnv(v)
			}
		})
	}
}

func (p *Particle) populateEngineBinaries() {
	runnables := make(map[string]string)
	for _, v := range EngineCache {
		for bin, pathw := range v.Runnables {
			runnables[bin] = pathw
			sym := filepath.Join(p.dir, "bin", bin+utils.SymlinkPostfix)
			if _, err := os.Stat(sym); err == nil {
				os.Remove(sym)
			}
			err := os.Symlink(pathw, filepath.Join(p.dir, "bin", bin+utils.SymlinkPostfix))

			if err != nil {
				fmt.Println(color.RedString("\nERROR: "), err)
				panic(err)
			}
		}
	}
	d, _ := json.MarshalIndent(runnables, "", "\t")
	_ = os.WriteFile(filepath.Join(p.dir, "engines", "run.json"), d, 0755)
	fmt.Println(color.GreenString("Installed %d engines and exposed %d binaries", len(EngineCache), len(runnables)))
}

func (p *Particle) prepareApplicable() {
	for _, app := range p.layers {
		if app == nil {
			continue
		}

		if app.base != nil {
			app.base.prepareBase(filepath.Join("src", app.Manifest.Name))
		}

		progress := utils.NewTreeProgress()
		progress.TrackFunction(fmt.Sprintf("→ %s [%s]...", app.Manifest.Name, app.Manifest.Block), func() {
			layer, ok := LayerCache[app.Manifest.Block]
			if !ok {
				fmt.Println(color.RedString("\nERROR: "), app.Manifest.Block+": layer was not found. Not sure what to do...")
				panic(ok)
			}
			err := layer.ExtractTo(filepath.Join(p.dir, "src", app.Manifest.Name))
			if err != nil {
				fmt.Println(color.RedString("\nERROR: "), err)
				panic(err)
			}
			for k, v := range app.Manifest.Meta {
				MetaCache[k] = os.ExpandEnv(v)
			}
		})
	}
}

func (p *Particle) executeApplicable() {
	for _, app := range p.layers {
		if app == nil {
			continue
		}
		app.executeApplicable()
		fmt.Println(color.GreenString("→ %s [%s]...", p.Manifest.Name, p.Manifest.Block))
		for _, ex := range app.Manifest.Recipe.Run {
			cmd := PrepareExecutor(p.dir, ex, filepath.Join("src", app.Manifest.Name))
			err := cmd.Run()
			if err != nil {
				fmt.Println(color.RedString("\nERROR: "), err)
				panic(err)
			}
		}
	}
}

func (p *Particle) calculateIntegrityHash() {
	progress := utils.NewTreeProgress()

	progress.TrackFunction("Calculating integrity hash", func() {
		hashes := p.getIntegrityHashInternal()
		d, err := json.MarshalIndent(hashes, "", "\t")
		if err != nil || os.WriteFile(filepath.Join(p.dir, "integrity.json"), d, 0755) != nil {
			fmt.Println(color.RedString("\nERROR: "), "Unable to generate JSON")
			panic(err)
		}
	})
}

func (p *Particle) getIntegrityHashInternal() map[string]string {
	hashes := make(map[string]string)
	fl := utils.LDir(filepath.Join(p.dir, "dist"), "")
	for _, f := range fl {
		hs, err := utils.CalcFileHash(filepath.Join(p.dir, "dist", f))
		if err != nil {
			fmt.Println(color.RedString("\nERROR: "), "Unable to calculate hash for", f)
			panic(err)
		}
		hashes[f] = hs
	}
	return hashes
}

func (p *Particle) compareIntegrity() {
	newHashes := p.getIntegrityHashInternal()
	oldHashes := make(map[string]string)
	deltas := make([]string, 1)
	deletions := make([]string, 1)
	snew := 0
	srem := 0
	d, err := os.ReadFile(filepath.Join(p.dir, "integrity.json"))
	if err != nil || json.Unmarshal(d, &oldHashes) != nil {
		fmt.Println(color.RedString("\nERROR: "), "Unable to read integrity.json")
		panic(err)
	}
	for f, hNew := range newHashes {
		hOld, ok := oldHashes[f]
		if !ok || hNew != hOld {
			fmt.Println(color.GreenString("+ %s [%s]", f, hNew))
			snew++
			deltas = append(deltas, f)
			os.MkdirAll(filepath.Dir(filepath.Join(p.dir, "out", f)), 0755)
			fh, err1 := os.Open(filepath.Join(p.dir, "dist", f))
			nh, err2 := os.Create(filepath.Join(p.dir, "out", f))
			if err1 != nil || err2 != nil {
				fmt.Println(color.RedString("\nERROR: "), err1, err2)
				panic(err)
			}
			defer fh.Close()
			defer nh.Close()
			_, err3 := io.Copy(nh, fh)
			if err3 != nil {
				fmt.Println(color.RedString("\nERROR: "), err3)
				panic(err)
			}
		}
	}
	for f, hOld := range oldHashes {
		_, ok := newHashes[f]
		if !ok {
			fmt.Println(color.RedString("- %s [%s]", f, hOld))
			srem++
			deletions = append(deletions, f)
		}
	}
	if len(deletions) > 0 {
		deletionsData := strings.Join(deletions, "\n")
		err = os.WriteFile(filepath.Join(p.dir, "out", ".deletions"), []byte(deletionsData), 0755)
		if err != nil {
			fmt.Println(color.RedString("\nERROR: "), err)
			panic(err)
		}
	}
}

func (p *Particle) fetchManifest(pkg string) (*structs.Manifest, error) {
	r, err := http.Get(p.Manifest.GetServer() + path.Join("repo", pkg, "particle.json"))
	if err != nil {
		return nil, err
	}
	if r.StatusCode != 200 {
		return nil, errors.New(r.Status)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	manifest := new(structs.Manifest)
	err = json.Unmarshal(body, &manifest)
	return manifest, err
}

func (p *Particle) loadParticle(pkg string) (particle *Particle, err error) {
	if pkg == "blank" {
		return nil, err
	}
	particle, ok := ParticleCache[pkg]
	if ok {
		return particle, nil
	}
	manif, err := p.fetchManifest(pkg)
	if err != nil {
		return nil, err
	}
	particle = &Particle{Manifest: *manif, dir: p.dir}
	ParticleCache[pkg] = particle
	return particle, nil
}
