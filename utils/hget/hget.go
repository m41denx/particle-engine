package hget

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var HGET_PREFIX = "."

var DisplayProgress = true

func Execute(url string, state *State, threads int, skiptls bool) (err error) {
	//otherwise is hget <url> command

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	//set up parallel

	var files = make([]string, 0)
	var parts = make([]Part, 0)
	var isInterrupted = false

	doneChan := make(chan bool, threads)
	fileChan := make(chan string, threads)
	errorChan := make(chan error, 1)
	stateChan := make(chan Part, 1)
	interruptChan := make(chan bool, threads)

	var downloader *HttpDownloader
	if state == nil {
		downloader = NewHttpDownloader(url, threads, skiptls)
	} else {
		downloader = &HttpDownloader{url: state.Url, file: filepath.Base(state.Url), par: int64(len(state.Parts)), parts: state.Parts, resumable: true}
	}
	go downloader.Do(doneChan, fileChan, errorChan, interruptChan, stateChan)

	for {
		select {
		case <-signalChan:
			//send par number of interrupt for each routine
			isInterrupted = true
			for i := 0; i < threads; i++ {
				interruptChan <- true
			}
		case file := <-fileChan:
			files = append(files, file)
		case err = <-errorChan:
			Errorf("%v", err)
			panic(err) //maybe need better style
		case part := <-stateChan:
			parts = append(parts, part)
		case <-doneChan:
			if isInterrupted {
				if downloader.resumable {
					Printf("Interrupted, saving state ... \n")
					s := &State{Url: url, Parts: parts}
					err = s.Save()
					if err != nil {
						Errorf("%v\n", err)
					}
					return
				} else {
					Warnf("Interrupted, but downloading url is not resumable, silently die")
					return
				}
			} else {
				err = JoinFile(files, filepath.Join(HGET_PREFIX, filepath.Base(url)))
				FatalCheck(err)
				err = os.RemoveAll(FolderOf(url))
				FatalCheck(err)
				return
			}
		}
	}
}
