package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

type GoMetrics struct {
	startTime time.Time
	lastTime  time.Time
	steps     map[string]int64 // ms took
	lastStep  string
}

func NewGoMetrics() *GoMetrics {
	return &GoMetrics{
		startTime: time.Now(),
		lastTime:  time.Now(),
		steps:     make(map[string]int64),
	}
}

func (gm *GoMetrics) Reset() {
	gm.startTime = time.Now()
	gm.lastTime = time.Now()
}

func (gm *GoMetrics) NewStep(name string) {
	gm.steps[name] = 0
	t := time.Now()
	if len(gm.lastStep) != 0 {
		gm.steps[gm.lastStep] = t.Sub(gm.lastTime).Milliseconds()
	}
	gm.lastStep = name
	gm.lastTime = t // Update total time
}

func (gm *GoMetrics) ExplicitDoneStep(name string) {
	gm.steps[name] = time.Now().Sub(gm.lastTime).Milliseconds()
	gm.lastStep = ""
	gm.lastTime = time.Now()
}

func (gm *GoMetrics) Done() {
	gm.steps[gm.lastStep] = time.Now().Sub(gm.lastTime).Milliseconds()
	gm.lastStep = ""
	gm.lastTime = time.Now()
}

func (gm *GoMetrics) dumpText(del string) string {
	out := ""
	for k, v := range gm.steps {
		out += fmt.Sprintf("%s: %dms"+del, k, v)
	}
	out += fmt.Sprintf("Total: %dms"+del, gm.lastTime.Sub(gm.startTime).Milliseconds())
	return out
}

func (gm *GoMetrics) DumpText() string {
	return gm.dumpText("\n")
}

func (gm *GoMetrics) DumpTextInline() string {
	return gm.dumpText("; ")
}

func (gm *GoMetrics) DumpJSON() string {
	// copy gm.stats
	data, _ := json.Marshal(struct {
		Steps map[string]int64
		Total int64
	}{
		Steps: gm.steps,
		Total: gm.lastTime.Sub(gm.startTime).Milliseconds(),
	})
	return string(data)
}
