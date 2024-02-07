package downloader

import (
	"github.com/cheggaaa/pb/v3"
	"runtime"
	"sync"
)

type Downloader struct {
	pool    *pb.Pool
	showBar bool
	threads int
	retries int
	jobs    []*Job
	wg      *sync.WaitGroup
}

func NewDownloader(threads int) *Downloader {
	if threads == 0 {
		threads = runtime.NumCPU()
	}
	return &Downloader{threads: threads, retries: 3, pool: pb.NewPool()}
}

func (d *Downloader) ShowBar(show bool) {
	d.showBar = show
}

func (d *Downloader) SetRetries(retries int) {
	d.retries = retries
}

func (d *Downloader) AddJob(job *Job) {
	if d.showBar {
		sz, _ := job.PrefetchSize()
		bar := pb.New64(sz)
		bar.SetTemplateString(`{{with string . "prefix"}}{{.}} {{end}}{{counters . }} {{bar . "[" "—" "›" "•" "]"}} {{with string . "suffix"}} {{.}}{{end}}`)
		//bar.SetTemplateString(`{{with string . "prefix"}}{{.}} {{end}}{{counters . }} {{bar . "[" "—" "›" "•" "]"}} {{speed . "%s/s" }} {{percent . }}{{with string . "suffix"}} {{.}}{{end}}`)
		job = job.WithBar(bar)
		d.pool.Add(bar)
	}
	job.SetRetries(d.retries)
	d.jobs = append(d.jobs, job)
}

func (d *Downloader) Do() []error {
	d.wg = new(sync.WaitGroup)
	d.wg.Add(len(d.jobs))

	var errs []error
	for _, job := range d.jobs {
		job := job
		go func() {
			err := job.Do(d.wg)
			if err != nil {
				errs = append(errs, err)
			}
		}()
	}
	err := d.pool.Start()
	if err != nil {
		errs = append(errs, err)
	}
	d.wg.Wait()
	return errs
}
