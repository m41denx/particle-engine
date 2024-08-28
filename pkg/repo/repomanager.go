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
	"path"
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

func (r *RepoManager) Pull(name string) error {
	panic("not implemented")
	return nil
	//	var err error
	//
	//	if r.arch == "" {
	//		r.arch = utils.GetArchString()
	//	}
	//
	//	// Parse versions
	//	if len(r.meta.Fullname) == 0 {
	//		r.meta, err = manifest.ParseParticleURL(name)
	//	} else {
	//		r.meta, err = manifest.ParseParticleURL(r.name)
	//	}
	//	if err != nil {
	//		return err
	//	}
	//
	//	if len(r.version) == 0 {
	//		ver := strings.SplitN(name, "@", 2)
	//		if len(ver) == 2 {
	//			r.version = ver[1]
	//		} else {
	//			r.version = "latest"
	//		}
	//	} else {
	//		r.version = strings.SplitN(r.version, "@", 2)[0]
	//	}
	//
	//	name = r.name + "@" + r.version
	//
	//	virtparticle, err := NewParticleFromString(fmt.Sprintf(`
	//{
	//	"name": "blank",
	//	"server": "%s",
	//	"recipe": {
	//		"base": "%s",
	//		"apply": [],
	//		"engines": [],
	//		"run": []
	//	}
	//
	//}`, r.url, name))
	//
	//	if err != nil {
	//		return err
	//	}
	//	virtparticle.Analyze(true)
	//	return nil
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
		_, err := os.Stat(path.Join(ctx.GetBuildDir(), "layers", manif.Layer.Block))
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
			uploadURL, nil, "layer", path.Join(ctx.GetBuildDir(), "layers", manif.Layer.Block),
		)
		if err != nil {
			return err
		}
		req, _ = pkg.Config.GenerateRequestURL(req, r.url)

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
