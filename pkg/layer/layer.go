package layer

import (
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg"
	"github.com/m41denx/particle-engine/utils"
	"github.com/m41denx/particle-engine/utils/downloader"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const DefaultLayerRepo = "http://particles.fruitspace.one/layers/"

type Layer struct {
	Hash      string
	Files     []string
	Deletions []string
	Size      int64

	//For download
	dir      string
	filename string
	server   string
}

func NewLayer(hash string, dir string, server string) *Layer {
	if server == "" {
		server = DefaultLayerRepo
	}
	return &Layer{
		Hash:      hash,
		Files:     []string{},
		Deletions: []string{},
		dir:       dir,
		filename:  path.Join(dir, "layers", hash+".7z"),
		server:    server,
	}
}

func CreateLayerFrom(dir string, blankLayer *Layer) (*Layer, error) {
	tempFile := path.Join(os.TempDir(), "_pbuild_"+time.Now().Format("20060102150405")+".7z")
	defer os.Remove(tempFile)
	err := pkg.UnzipProvider.OpenZip(path.Join(os.TempDir(), tempFile)).WorkDir(dir).AddDirectory("").Compress()
	if err != nil {
		return nil, err
	}
	hash, err := utils.CalcFileHash(tempFile)
	if err != nil {
		return nil, err
	}
	layer := NewLayer(hash, blankLayer.dir, blankLayer.server)
	return layer, os.Rename(tempFile, layer.filename)
}

func (l *Layer) Download(dlmgr *downloader.Downloader) error {
	if l.isLocalCopyValid() {
		return nil
	}
	l.Size = 0
	httpClient := GetLayerFetcher()
	job := downloader.NewJob(l.server+l.Hash, "GET", l.filename)
	job.SetHttpClient(httpClient)
	job.WithLabel(color.GreenString("→ %s", l.Hash))
	dlmgr.AddJob(job)
	return nil
}

func (l *Layer) ExtractTo(dest string) (err error) {
	err = pkg.UnzipProvider.OpenZip(l.filename).Decompress(dest)
	if err != nil {
		return err
	}

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

func (l *Layer) isLocalCopyValid() bool {
	var stat syscall.Stat_t
	err := syscall.Stat(l.filename, &stat)
	if err != nil {
		return false
	}
	l.Size = stat.Size
	hash, err := utils.CalcFileHash(l.filename)
	if err != nil {
		return false
	}
	return hash == l.Hash
}
