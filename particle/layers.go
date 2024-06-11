package particle

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/utils"
	"github.com/m41denx/particle/utils/downloader"
	"os"
	"path"
	"path/filepath"
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

func (l *Layer) Fetch(id string, pname string, dlmgr *downloader.Downloader) error {
	l.ID = id
	f, err := os.Stat(path.Join(l.d, l.ID))
	if err != nil {
		l.ScheduleDownload(l.ID, pname, dlmgr)
	} else {
		fmt.Print(color.GreenString("→ %s [%s]", pname, id))
		sz := float64(f.Size()) / (1024 * 1024) //MiB
		if sz > 1024 {
			fmt.Println(color.GreenString(" • %.1f GB", sz/1024))
		} else {
			fmt.Println(color.GreenString(" • %.1f MB", sz))
		}
	}
	return nil
}

func (l *Layer) MatchHashes() error {
	hs, err := l.CalcHash()
	if err != nil {
		return err
	}
	if hs != l.ID {
		info, _ := os.Stat(path.Join(l.d, l.ID))
		sz := float64(info.Size()) / (1024 * 1024) //MiB
		os.Remove(path.Join(l.d, l.ID))            // Remove layer so it can be re-downloaded
		return errors.New(fmt.Sprintf("MD5 Hashes don't match: %s and %s (%.2f MB)", l.ID, hs, sz))
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

func (l *Layer) ScheduleDownload(id string, pname string, dlmgr *downloader.Downloader) {
	job := downloader.NewJob(l.server+path.Join("/layers", id), "GET", path.Join(l.d, id))
	job.WithLabel(color.GreenString("→ %s [%s]", pname, id))
	dlmgr.AddJob(job)
}
