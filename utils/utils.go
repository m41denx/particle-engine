package utils

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/m41denx/particle/structs"
	"io"
	"os"
	"path"
	"runtime"
)

const GitIgnore = `bin/
dist/
engines/
out/
src/`

func CalcFileHash(dp string) (string, error) {
	hash := md5.New()
	fp, err := os.Open(dp)
	defer fp.Close()
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(hash, fp); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func LDir(srcDir string, prefix string) []string {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return nil
	}
	var dlist []string
	for _, file := range files {
		pr := path.Join(srcDir, file.Name())
		if file.IsDir() {
			flist := LDir(pr, prefix+file.Name()+"/")
			dlist = append(dlist, flist...)
		} else {
			// file
			p := prefix + file.Name()
			dlist = append(dlist, p)
		}
	}
	return dlist
}

func GetArchString() string {
	var arch string
	switch runtime.GOOS {
	case "windows":
		arch += "w"
	case "linux":
		arch += "l"
	case "darwin":
		arch += "d"
	default:
		return "unsupported"
	}
	switch runtime.GOARCH {
	case "amd64":
		arch += "64"
	case "386":
		arch += "32"
	case "arm64":
		arch += "64a"
	default:
		return "unsupported"
	}
	return arch
}

func PrepareStorage() string {
	home, _ := os.UserCacheDir()
	pc := path.Join(home, "particle_cache")
	os.MkdirAll(pc, 0750)
	os.MkdirAll(path.Join(pc, "layers"), 0750)
	os.MkdirAll(path.Join(pc, "repo"), 0750)
	return pc
}

func ParticleInit(pathname string) {
	if pathname == "" {
		pathname = "."
	}
	os.MkdirAll(pathname, 0750)
	os.Chdir(pathname)
	manifest, _ := json.MarshalIndent(structs.NewManifest(), "", "\t")
	os.WriteFile("particle.json", manifest, 0750)
	os.MkdirAll("out", 0750)
	os.MkdirAll("bin", 0750)
	os.MkdirAll("dist", 0750)
	os.MkdirAll("src", 0750)
	os.MkdirAll("engines", 0750)
	os.WriteFile(".gitignore", []byte(GitIgnore), 0750)
	Log("Init done at", pathname)
}
