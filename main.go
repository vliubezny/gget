package main

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

const usage = "usage:  gget fileURL ..."

func main() {
	urls := os.Args[1:]
	if len(urls) == 0 {
		terminate(usage)
	}

	for _, fileURL := range urls {
		err := validateURL(fileURL)
		if err != nil {
			terminate(err.Error())
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		terminate("Canceled")
	}()

	progressChannels := make(map[string]<-chan float64)
	for _, fileURL := range urls {
		filePath := filepath.Base(fileURL)
		progressChannels[filePath] = downloadFile(filePath, fileURL)
	}

	initWriter(os.Stdout)
	err := observeStatus(progressChannels)
	if err != nil {
		terminate(err.Error())
	}
}

func validateURL(fileURL string) error {
	u, err := url.Parse(fileURL)
	if err != nil {
		return err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL should have http(s) scheme: %s", fileURL)
	}

	if len(u.EscapedPath()) <= 1 {
		return fmt.Errorf("URL path is missing: %s", fileURL)
	}

	return nil
}

func terminate(message string) {
	println(message)
	os.Exit(1)
}
