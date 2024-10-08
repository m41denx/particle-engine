package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/minio/selfupdate"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var SUPPORTED_ARCH = []string{"w64", "l64", "l64a", "d64", "d64a"}

func CalcFileHash(dp string) (string, error) {
	hash := sha256.New()
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

func CalcHash(b []byte) string {
	hash := sha256.New()
	hash.Write(b)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func LDir(srcDir string, prefix string) []string {
	files, err := os.ReadDir(filepath.Clean(srcDir))
	if err != nil {
		return nil
	}
	var dlist []string
	for _, file := range files {
		pr := filepath.Join(srcDir, file.Name())
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

	if archOverride, ok := os.LookupEnv("PARTICLE_ARCH"); ok {
		arch = strings.ToLower(archOverride)
	}
	return arch
}

func PrepareStorage() string {
	home, _ := os.UserCacheDir()
	pc := filepath.Join(home, "particle_cache")
	_ = os.MkdirAll(pc, 0750)
	_ = os.MkdirAll(filepath.Join(pc, "layers"), 0750)
	_ = os.MkdirAll(filepath.Join(pc, "repo"), 0750)
	_ = os.MkdirAll(filepath.Join(pc, "temp"), 0750)
	return pc
}

func GetUpdate(ver string, update bool) error {
	ver = "v" + ver
	res, err := http.Get("https://api.github.com/repos/m41denx/particle-engine/releases/latest")
	if err != nil {
		return err
	}
	d, _ := io.ReadAll(res.Body)
	var apiresponse map[string]interface{}
	err = json.Unmarshal(d, &apiresponse)
	if err != nil {
		return err
	}
	newver := apiresponse["tag_name"].(string)
	if ver == newver {
		fmt.Println(color.GreenString("[UPD] Already up to date: %s", ver))
		return nil
	}
	fmt.Println(color.YellowString("[UPD] Found new version %s (Current: %s)", newver, ver))
	changelog := strings.Split(apiresponse["body"].(string), "\n")
	for _, l := range changelog {
		l = strings.TrimSpace(l)
		l = strings.TrimRight(l, "\r\n")
		if l[0] == '#' {
			fmt.Println("===" + color.CyanString(strings.TrimLeft(l, "#")))
			continue
		}
		if l[0] == '*' || l[0] == '-' {
			fmt.Println(color.GreenString("*" + l[1:]))
		}
	}
	var downloadlink string
	for _, asset := range apiresponse["assets"].([]interface{}) {
		asset := asset.(map[string]interface{})
		if strings.Contains(asset["name"].(string), runtime.GOOS) &&
			strings.Contains(asset["name"].(string), runtime.GOARCH) {
			downloadlink = asset["browser_download_url"].(string)
			break
		}
	}
	if downloadlink == "" {
		return fmt.Errorf("No download link for your architecture found")
	}
	if update {
		return SelfUpdate(downloadlink)
	}
	return nil
}

func SelfUpdate(url string) error {
	progress := NewTreeProgress()
	var wg sync.WaitGroup
	wg.Add(1)
	c := make(chan bool)
	go progress.Run("Downloading update for "+GetArchString(), &wg, c)
	dat, err := http.Get(url)
	if err != nil {
		c <- true
		return err
	}
	err = selfupdate.Apply(dat.Body, selfupdate.Options{})
	c <- true
	wg.Wait()
	return err
}

func MapDiff(old map[string]string, new map[string]string) (additions map[string]string, deletions []string) {

	for file, newHash := range new {
		oldHash, ok := old[file]
		if ok && newHash == oldHash {
			continue
		}
		additions[file] = newHash
	}

	for file, _ := range old {
		_, ok := new[file]
		if ok {
			continue
		}
		deletions = append(deletions, file)
	}

	return
}
