package utils

import (
	"fmt"
	"sync"
	"time"
)

type TreeProgress struct {
	indent      int
	loadsyms    []string
	deltaTimeMs int
	seq         int

	dl int
}

func NewTreeProgress() *TreeProgress {
	return &TreeProgress{
		loadsyms:    []string{"⣷", "⣯", "⣟", "⡿", "⢿", "⣻", "⣽", "⣾"},
		deltaTimeMs: 100,
	}
}

func (t *TreeProgress) Tab() {
	t.indent++
}
func (t *TreeProgress) Ret() {
	t.indent--
	if t.indent < 0 {
		t.indent = 0
	}
}

func (t *TreeProgress) Run(text string, wg *sync.WaitGroup, wait chan bool) {
	for {
		select {
		case <-wait:
			u := "\r"
			for i := 0; i < t.dl; i++ {
				u += " "
			}
			fmt.Print(u)
			fmt.Printf("\r%s%s %s\n", t.genIndent(), text, "✔")
			wg.Done()
			return
		default:
			v := fmt.Sprintf("\r%s%s %s\r", t.genIndent(), text, t.loadsyms[t.seq])
			t.dl = len(v)
			fmt.Print(v)
			t.seq++
			if t.seq >= len(t.loadsyms) {
				t.seq = 0
			}
			dur := time.Duration(int(time.Millisecond) * t.deltaTimeMs)
			time.Sleep(dur)
		}
	}
}

func (t *TreeProgress) TrackFunction(text string, f func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	c := make(chan bool)
	go t.Run(text, &wg, c)
	f()
	c <- true
	wg.Wait()
}

func (t *TreeProgress) genIndent() string {
	s := ""
	for i := 0; i < t.indent; i++ {
		s += "•  "
	}
	return s
}
