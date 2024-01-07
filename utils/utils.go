package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
)

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
