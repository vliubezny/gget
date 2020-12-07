package main

import (
	"fmt"
	"io"
	"sync"
	"text/tabwriter"
	"time"
)

var tw *tabwriter.Writer

func initWriter(output io.Writer) {
	tw = tabwriter.NewWriter(output, 10, 1, 2, ' ', 0)
}

func observeStatus(progressChannels map[string]<-chan float64) error {
	wg := &sync.WaitGroup{}
	wg.Add(len(progressChannels))
	state := &sync.Map{}

	// store headers to preserve column ordering
	headers := make([]string, 0, len(progressChannels))

	for filePath, progressCh := range progressChannels {
		headers = append(headers, filePath)
		state.Store(filePath, 0.0)
		go func(filePath string, progressCh <-chan float64) {
			defer wg.Done()
			for progress := range progressCh {
				state.Store(filePath, progress)
			}
		}(filePath, progressCh)
	}

	printTable(headers, state)

	done := make(chan bool, 1)
	go func() {
		wg.Wait()
		done <- true
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			moveCursorBeforeTable()
			if err := printTable(headers, state); err != nil {
				return err
			}
		case <-done:
			moveCursorBeforeTable()
			return printTable(headers, state)
		}
	}
}

func moveCursorBeforeTable() {
	fmt.Print("\033[2A")
}

func printTable(headers []string, state *sync.Map) error {
	for _, filePath := range headers {
		fmt.Fprintf(tw, "%s\t", filePath)
	}
	fmt.Fprintln(tw)

	for _, filePath := range headers {
		val, _ := state.Load(filePath)
		progress := val.(float64)
		switch {
		case progress >= 0:
			fmt.Fprintf(tw, "%6.2f%%\t", progress)
		case progress == ProgressNotAvailable:
			fmt.Fprintf(tw, "N/A\t")
		case progress == DownloadFailure:
			fmt.Fprintf(tw, "Error\t")
		default:
			fmt.Fprintf(tw, "Error: %.0f\t", progress)
		}
	}
	fmt.Fprintln(tw)

	return tw.Flush()
}
