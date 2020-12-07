package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

const usage = "usage:  gget fileURL ..."

func main() {
	log.SetFlags(0) // disable timestamps

	urls := os.Args[1:]
	if len(urls) == 0 {
		log.Fatal(usage)
	}

	for _, fileURL := range urls {
		err := validateURL(fileURL)
		if err != nil {
			log.Fatal(err)
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Fatal("Canceled")
	}()

	progressChannels := make(map[string]<-chan float64)
	for _, fileURL := range urls {
		filePath := filepath.Base(fileURL)
		progressChannels[filePath] = downloadFile(filePath, fileURL)
	}

	initWriter(os.Stdout)
	err := observeStatus(progressChannels)
	if err != nil {
		log.Fatal(err.Error())
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

	return nil
}
