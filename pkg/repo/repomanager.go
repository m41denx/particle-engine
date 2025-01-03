package particle

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"github.com/m41denx/particle-engine/pkg"
	"github.com/m41denx/particle-engine/pkg/builder"
	"github.com/m41denx/particle-engine/pkg/manifest"
	"github.com/m41denx/particle-engine/utils"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func NewRepoMgr() *RepoManager {
	return &RepoManager{
		url: manifest.DefaultRepo,
	}
}

type RepoManager struct {
	url      string
	name     string
	meta     manifest.ParticleMeta
	version  string
	private  bool
	unlisted bool
	arch     string
}

func (r *RepoManager) WithUrl(url string) *RepoManager {
	if len(url) > 0 {
		r.url = url
	}
	return r
}

func (r *RepoManager) WithName(name string) *RepoManager {
	r.name = name
	return r
}

func (r *RepoManager) WithVersion(version string) *RepoManager {
	r.version = version
	return r
}

func (r *RepoManager) WithArch(arch string) *RepoManager {
	r.arch = arch
	return r
}

func (r *RepoManager) WithPrivate(private bool) *RepoManager {
	r.private = private
	return r
}

func (r *RepoManager) WithUnlisted(unlisted bool) *RepoManager {
	r.unlisted = unlisted
	return r
}

func (r *RepoManager) Publish(ctx *builder.BuildContext) error {
	manif := ctx.Manifest

	if r.arch == "" {
		r.arch = utils.GetArchString()
	}

	// Parse versions
	if len(r.name) == 0 {
		r.name = strings.SplitN(manif.Name, "@", 2)[0]
	}
	if !strings.Contains(r.name, "/") {
		uname, _ := pkg.Config.GetCredsForRepo(r.url)
		r.name = fmt.Sprintf("%s/%s", uname, r.name)
	}
	if len(r.version) == 0 {
		ver := strings.SplitN(manif.Name, "@", 2)
		if len(ver) == 2 {
			r.version = ver[1]
		} else {
			r.version = "latest"
		}
	}

	link, err := url.JoinPath(r.url, fmt.Sprintf("%s@%s", r.name, r.version))

	if err != nil {
		return err
	}
	r.meta, err = manifest.ParseParticleURL(link)

	if err != nil {
		return err
	}

	{
		// Check particle existence
		_, err := os.Stat(filepath.Join(ctx.GetHomeDir(), "layers", manif.Layer.Block+".7z"))
		if err != nil {
			fmt.Println(color.RedString("\nError reading layer for "+r.meta.Name+": "), "\nPlease run particle build first")
			return err
		}
	}

	prg := utils.NewTreeProgress()

	// Upload manifest

	err = prg.TrackFunction(color.BlueString("Pushing manifest for %s...", r.meta.Fullname), func() error {
		mdata := manif.ToJson()
		uploadURL, err := url.JoinPath(r.url, "upload", r.meta.Fullname, r.arch)
		if err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodPost, uploadURL, strings.NewReader(mdata))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req, _ = pkg.Config.GenerateRequestURL(req, r.url)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			errd, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("Failed to push manifest: %d\n%s", resp.StatusCode, string(errd))
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Upload layer

	err = prg.TrackFunction(color.BlueString("Pushing layer %s...", manif.Layer.Block[:12]), func() error {
		uploadURL, err := url.JoinPath(r.url, "upload", r.meta.Fullname, r.arch)
		if err != nil {
			return err
		}
		req, err := newfileUploadRequest(
			uploadURL, nil, "layer", filepath.Join(ctx.GetHomeDir(), "layers", manif.Layer.Block+".7z"),
		)
		if err != nil {
			return err
		}
		req, _ = pkg.Config.GenerateRequestURL(req, r.url)

		req.Header.Set("Layer-Hash", manif.Layer.Block)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			errd, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("Failed to push layer: %d\n%s", resp.StatusCode, string(errd))
		}
		return nil
	})

	return err
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
