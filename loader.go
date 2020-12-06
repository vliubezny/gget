package main

import (
	"io"
	"net/http"
	"os"
	"time"
)

const (
	ProgressNotAvailable = -1
	DownloadFailure      = -2
)

var client = http.Client{
	Timeout: 10 * time.Minute,
}

type progressNotifier struct {
	total      int64
	counter    int64
	progressCh chan<- float64
}

func (pn *progressNotifier) Write(p []byte) (n int, err error) {
	if pn.total > 0 {
		pn.counter += int64(len(p))
		var progress float64 = float64(pn.counter) * 100 / float64(pn.total)
		pn.progressCh <- progress
	} else {
		pn.progressCh <- ProgressNotAvailable
	}
	return len(p), nil
}

func downloadFile(filePath, fileURL string) <-chan float64 {
	output := make(chan float64)
	go func() {
		defer close(output)
		resp, err := client.Get(fileURL)
		if err != nil {
			output <- DownloadFailure
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			output <- DownloadFailure
			return
		}

		out, err := os.Create(filePath)
		if err != nil {
			output <- DownloadFailure
			return
		}
		defer out.Close()

		notifier := &progressNotifier{
			total:      resp.ContentLength,
			progressCh: output,
		}
		tee := io.TeeReader(resp.Body, notifier)

		_, err = io.Copy(out, tee)
		if err != nil {
			output <- DownloadFailure
		}
	}()
	return output
}
