package downloader

import (
	"errors"
	"github.com/cheggaaa/pb/v3"
	"io"
	"net/http"
	"os"
	"sync"
)

type Job struct {
	url      string
	method   string
	progress *pb.ProgressBar
	retries  int
	fp       string
	showBar  bool
	label    string
}

func NewJob(url string, method string, fp string) *Job {
	return &Job{url: url, method: method, retries: 3, fp: fp}
}

func (j *Job) SetRetries(retries int) {
	j.retries = retries
}

func (j *Job) WithBar(bar *pb.ProgressBar) *Job {
	j.showBar = true
	j.progress = bar
	return j
}

func (j *Job) WithLabel(label string) *Job {
	j.label = label
	return j
}

func (j *Job) PrefetchSize() (sz int64, err error) {
	req, err := http.Head(j.url)
	if err != nil {
		return 0, err
	}
	if req.StatusCode >= 400 {
		return 0, errors.New(req.Status)
	}
	return req.ContentLength, nil
}

func (j *Job) Do(wg *sync.WaitGroup) error {
	defer wg.Done()
	req, err := http.NewRequest(j.method, j.url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if j.retries > 0 {
			j.retries--
			return j.Do(wg)
		}
		return err
	}
	defer resp.Body.Close()
	if j.showBar {
		j.progress.Set(pb.Bytes, true)
		j.progress.Set(pb.SIBytesPrefix, true)
		j.progress.Set("prefix", j.label)
	}
	src := resp.Body
	if j.showBar {
		src = j.progress.NewProxyReader(src)
	}

	dst, err := os.Create(j.fp)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if j.showBar {
		j.progress.Finish()
	}
	return err
}
