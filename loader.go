package main

import (
	"io"
	"net/http"
	"os"
	"time"
)

const (
	ProgressNotAvailable = -1.0
	DownloadFailure      = -2.0
)

var client = http.Client{
	Timeout: 10 * time.Minute,
}

type progressNotifier struct {
	total      int64
	processed  int64
	progressCh chan<- float64
}

func (pn *progressNotifier) Write(p []byte) (n int, err error) {
	pn.processed += int64(len(p))
	var progress float64 = float64(pn.processed) * 100 / float64(pn.total)
	pn.progressCh <- progress
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

		var source io.Reader
		if resp.ContentLength > 0 {
			notifier := &progressNotifier{
				total:      resp.ContentLength,
				progressCh: output,
			}
			source = io.TeeReader(resp.Body, notifier)
		} else {
			output <- ProgressNotAvailable
			source = resp.Body
		}

		if _, err = io.Copy(out, source); err != nil {
			output <- DownloadFailure
		}
	}()
	return output
}
