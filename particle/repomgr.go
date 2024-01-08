package particle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle/structs"
	"github.com/m41denx/particle/utils"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

func NewRepoMgr() *RepoMgr {
	return &RepoMgr{
		url: structs.DefaultRepo,
	}
}

type RepoMgr struct {
	url      string
	name     string
	version  string
	private  bool
	unlisted bool
}

func (r *RepoMgr) WithUrl(url string) *RepoMgr {
	if len(url) > 0 {
		r.url = url
	}
	return r
}

func (r *RepoMgr) WithName(name string) *RepoMgr {
	r.name = name
	return r
}

func (r *RepoMgr) WithVersion(version string) *RepoMgr {
	r.version = version
	return r
}

func (r *RepoMgr) WithPrivate(private bool) *RepoMgr {
	r.private = private
	return r
}

func (r *RepoMgr) WithUnlisted(unlisted bool) *RepoMgr {
	r.unlisted = unlisted
	return r
}

func (r *RepoMgr) Publish(p *Particle) error {
	m := p.Manifest
	var wg sync.WaitGroup

	// Parse versions
	if len(r.name) == 0 {
		r.name = strings.SplitN(m.Name, "@", 2)[0]
	} else {
		r.name = strings.SplitN(r.name, "@", 2)[0]
	}
	if len(r.version) == 0 {
		ver := strings.SplitN(m.Name, "@", 2)
		if len(ver) == 2 {
			r.version = ver[1]
		} else {
			r.version = "1.0"
		}
	} else {
		r.version = strings.SplitN(r.version, "@", 2)[0]
	}

	m.Name = r.name + "@" + r.version

	progress := utils.NewTreeProgress()

	log.Println(color.CyanString("Pushing %s manifest...", m.Name))

	progress.Tab()
	// Manifest
	{
		wg.Add(1)
		c := make(chan bool)
		go progress.Run("Pushing manifest for "+m.Name, &wg, c)

		mdata, _ := json.Marshal(m)

		req, err := http.NewRequest("POST", r.url+path.Join("/upload", m.Name, utils.GetArchString()), bytes.NewReader(mdata))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req, _ = Config.GenerateRequestURL(req, r.url)
		resp, err := http.DefaultClient.Do(req)
		c <- true
		wg.Wait()
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			errd, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("Failed to push manifest: %d\n%s", resp.StatusCode, string(errd))
		}
	}
	progress.Ret()

	progress.Tab()
	// Manifest
	{
		wg.Add(1)
		c := make(chan bool)
		go progress.Run("Pushing layer "+m.Block+" for "+m.Name, &wg, c)

		req, err := newfileUploadRequest(r.url+path.Join("/upload", m.Name, utils.GetArchString()),
			nil, "layer", p.dir+"/"+m.Block)
		req, _ = Config.GenerateRequestURL(req, r.url)

		if err != nil {
			log.Fatal(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)

		c <- true
		wg.Wait()
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			errd, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("Failed to push layer: %d\n%s", resp.StatusCode, string(errd))
		}
	}
	progress.Ret()

	return nil
}

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
