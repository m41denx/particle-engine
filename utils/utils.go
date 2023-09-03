package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
)

func CalcFileHash(dp string) (string, error) {
	hash := md5.New()
	fp, err := os.Open(dp)
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
