package particle

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/utils"
	"github.com/m41denx/particle/utils/hget"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type Layer struct {
	ID        string
	Files     []string
	Deletions []string
	Additions []string

	d      string
	server string
}

func NewLayer(server string) *Layer {
	home, _ := os.UserCacheDir()
	return &Layer{
		server: server,
		d:      filepath.Join(home, "particle_cache", "layers"),
	}
}

func (l *Layer) Fetch(id string) error {
	l.ID = id
	f, err := os.Stat(path.Join(l.d, l.ID))
	if err != nil {
		l.Download(l.ID)
	} else {
		sz := float64(f.Size()) / (1024 * 1024) //MiB
		if sz > 1024 {
			fmt.Println(color.GreenString(" • %.1f GB", sz/1024))
		} else {
			fmt.Println(color.GreenString(" • %.1f MB", sz))
		}
	}
	hs, err := l.CalcHash()
	if err != nil {
		return err
	}
	if hs != l.ID {
		return errors.New(fmt.Sprintf("MD5 Hashes don't match: %s and %s", l.ID, hs))
	}
	return nil
}

func (l *Layer) CalcHash() (string, error) {
	return utils.CalcFileHash(path.Join(l.d, l.ID))
}

func (l *Layer) ExtractTo(dest string) (err error) {
	//err = utils.Un7zip(path.Join(l.d, l.ID), dest)
	err = UnzipProvider.OpenZip(path.Join(l.d, l.ID)).Decompress(dest)
	if err != nil {
		return err
	}

	//Process deletions
	f, err := os.ReadFile(filepath.Join(dest, ".deletions"))
	if err == nil {
		l.Deletions = strings.Split(string(f), "\n")
		for i, d := range l.Deletions {
			if strings.TrimSpace(d) == "" {
				l.Deletions[i] = l.Deletions[len(l.Deletions)-1]
				l.Deletions = l.Deletions[:len(l.Deletions)-1]
			} else {
				err = os.Remove(filepath.Join(dest, d))
				if err != nil {
					return err
				}
			}
		}
	}
	os.Remove(filepath.Join(dest, ".deletions"))

	return nil
}

func (l *Layer) CreateLayer(from string, to string) (err error) {
	return UnzipProvider.OpenZip(to).WorkDir(from).AddDirectory("").Compress()
}

func (l *Layer) Download(id string) {
	numcpu := runtime.NumCPU()
	if numcpu > 8 {
		numcpu = 8
	}
	hget.HGET_PREFIX = l.d
	hget.Execute(l.server+path.Join("/layers", id), nil, numcpu, false)
}
