package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
)

func main() {
	urls := os.Args[1:]
	tw := tabwriter.NewWriter(os.Stdout, 0, 1, 3, ' ', 0)

	files := make([]string, len(urls))
	states := make([]string, len(urls))

	for i, u := range urls {
		fileName, err := getFileName(u)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		files[i] = fileName
		states[i] = "0%"
		fmt.Fprintf(tw, "%s\t", fileName)
	}
	fmt.Fprintln(tw)

	notifications := make(chan Notification)

	wc := &sync.WaitGroup{}
	for i, u := range urls {
		wc.Add(1)

		go func(jobID int, fileUrl string) {
			defer wc.Done()
			err := downloadFile(jobID, files[jobID], fileUrl, notifications)
			if err != nil {
				notifications <- Notification{
					jobID:  jobID,
					status: err.Error(),
				}
			}
		}(i, u)
	}

	go func() {
		wc.Wait()
		close(notifications)
	}()

	for n := range notifications {
		states[n.jobID] = n.status
		for _, s := range states {
			fmt.Fprintf(tw, "%s\t", s)
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

type Notification struct {
	jobID  int
	status string
}

type ProgressNotifier struct {
	jobID      int
	total      int64
	counter    int64
	progressCh chan<- Notification
}

func (pn *ProgressNotifier) Write(p []byte) (n int, err error) {
	notification := Notification{jobID: pn.jobID}
	if pn.total > 0 {
		pn.counter += int64(len(p))
		progress := pn.counter * 100 / pn.total
		notification.status = fmt.Sprintf("%d%%", progress)
	} else {
		notification.status = "n/a"
	}
	pn.progressCh <- notification
	return len(p), nil
}

func downloadFile(jobID int, filepath, fileURL string, progressCh chan<- Notification) error {
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	notifier := &ProgressNotifier{
		jobID:      jobID,
		total:      resp.ContentLength,
		progressCh: progressCh,
	}
	tee := io.TeeReader(resp.Body, notifier)

	_, err = io.Copy(out, tee)
	return err
}

func getFileName(fileURL string) (string, error) {
	u, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("URL should have http(s) scheme: %s", fileURL)
	}

	segments := strings.Split(u.EscapedPath(), "/")
	fileName := segments[len(segments)-1]

	if fileName == "" {
		return "", fmt.Errorf("Cannot extract file name from path: %s", fileURL)
	}

	return fileName, nil
}
