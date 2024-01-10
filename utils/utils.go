package utils

import (
	"crypto/md5"
	"fmt"
	"github.com/fatih/color"
	"github.com/minio/selfupdate"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	pc := filepath.Join(home, "particle_cache")
	os.MkdirAll(pc, 0750)
	os.MkdirAll(filepath.Join(pc, "layers"), 0750)
	os.MkdirAll(filepath.Join(pc, "repo"), 0750)
	return pc
}

func SelfUpdate(ver string) error {
	link := "https://s3.m41den.com/particle_releases/"
	res, err := http.Get(link + "ver")
	if err != nil {
		return err
	}
	d, _ := io.ReadAll(res.Body)
	newver := strings.Trim(string(d), "\n")
	if ver == newver {
		fmt.Println(color.GreenString("[UPD] Already up to date: %s", ver))
		return nil
	}
	fmt.Println(color.YellowString("[UPD] Found new version %s (Current: %s)", newver, ver))

	progress := NewTreeProgress()
	var wg sync.WaitGroup
	wg.Add(1)
	c := make(chan bool)
	go progress.Run("Downloading update for "+GetArchString(), &wg, c)
	dat, err := http.Get(link + "particle-" + GetArchString())
	if err != nil {
		c <- true
		return err
	}
	err = selfupdate.Apply(dat.Body, selfupdate.Options{})
	c <- true
	wg.Wait()
	return err
}
