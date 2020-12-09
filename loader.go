package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	progressNotAvailable = -1.0
	downloadFailure      = -2.0
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
	pn.progressCh <- float64(pn.processed) * 100 / float64(pn.total)
	return len(p), nil
}

func downloadFile(ctx context.Context, filePath, fileURL string) <-chan float64 {
	output := make(chan float64)
	go func() {
		defer close(output)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
		if err != nil {
			output <- downloadFailure
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			output <- downloadFailure
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			output <- downloadFailure
			return
		}

		out, err := os.Create(filePath)
		if err != nil {
			output <- downloadFailure
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
			output <- progressNotAvailable
			source = resp.Body
		}

		if _, err = io.Copy(out, source); err != nil && err != context.Canceled {
			output <- downloadFailure
		}
	}()
	return output
}
