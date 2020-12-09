package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const usage = "usage:  gget fileURL ..."

func main() {
	log.SetFlags(0) // disable timestamps

	urls := os.Args[1:]
	if len(urls) == 0 {
		log.Fatal(usage)
	}

	files := make([]string, len(urls))
	for i, fileURL := range urls {
		err := validateURL(fileURL)
		if err != nil {
			log.Fatal(err)
		}
		files[i] = filepath.Base(fileURL)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	state := newProgressState(files, os.Stdout)
	wg := &sync.WaitGroup{}

	for i, fileURL := range urls {
		wg.Add(1)
		progressCh := downloadFile(ctx, files[i], fileURL)
		go func(header string) {
			defer wg.Done()
			for progress := range progressCh {
				state.update(header, progress)
			}
		}(files[i])
	}

	if err := state.print(); err != nil {
		log.Fatal(err)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ticker.C:
				if err := state.print(); err != nil {
					log.Fatal(err)
				}
			case <-sigs:
				cancel()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	wg.Wait()

	if ctx.Err() == context.Canceled {
		log.Fatal("Canceled")
	}

	cancel()
	if err := state.print(); err != nil {
		log.Fatal(err)
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
