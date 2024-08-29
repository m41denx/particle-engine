package downloader

import (
	"github.com/cheggaaa/pb/v3"
	"golang.org/x/term"
	"math"
	"os"
	"runtime"
	"sync"
)

type Downloader struct {
	pool    *pb.Pool
	showBar bool
	threads int
	retries int
	jobs    []*Job
}

func NewDownloader(threads int) *Downloader {
	if threads == 0 {
		threads = runtime.NumCPU()
	}
	return &Downloader{threads: threads, retries: 1, pool: pb.NewPool()}
}

func (d *Downloader) ShowBar(show bool) {
	d.showBar = show
}

func (d *Downloader) ShowBarAuto() {
	// Check if we have tty
	d.showBar = term.IsTerminal(int(os.Stdout.Fd()))
}

func (d *Downloader) SetRetries(retries int) {
	d.retries = retries
}

func (d *Downloader) AddJob(job *Job) {
	if d.showBar {
		sz, _ := job.PrefetchSize()
		bar := pb.New64(sz)
		bar.SetTemplateString(`{{with string . "prefix"}}{{.}} {{end}}{{counters . }} {{bar . "[" "━" "›" "•" "]"}} {{with string . "suffix"}} {{.}}{{end}}`)
		//bar.SetTemplate(pb.Full)
		job = job.WithBar(bar)
		d.pool.Add(bar)
	}
	job.SetRetries(d.retries)
	d.jobs = append(d.jobs, job)
}

func (d *Downloader) Do() []error {
	var errs []error

	step := int(math.Min(float64(d.threads), float64(len(d.jobs))))
	for i := 0; i < len(d.jobs); i += step {
		wg := new(sync.WaitGroup)
		wg.Add(step)
		go func() {
			if err := d.jobs[i].Do(wg); err != nil {
				errs = append(errs, err)
			}
		}()
		if d.showBar {
			err := d.pool.Start()
			if err != nil {
				errs = append(errs, err)
			}
		}
		wg.Wait()
	}

	return errs
}
