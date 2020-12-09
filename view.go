package main

import (
	"fmt"
	"io"
	"sync"
	"text/tabwriter"
)

type progressState struct {
	mux      *sync.Mutex
	progress map[string]float64
	headers  []string
	writer   *tabwriter.Writer
	printed  bool
}

func newProgressState(headers []string, output io.Writer) *progressState {
	s := &progressState{
		mux:      &sync.Mutex{},
		progress: make(map[string]float64, len(headers)),
		headers:  headers,
		writer:   tabwriter.NewWriter(output, 10, 1, 2, ' ', 0),
	}

	for _, h := range headers {
		s.progress[h] = 0.0
	}

	return s
}

func (s *progressState) update(header string, progress float64) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.progress[header] = progress
}

func (s *progressState) print() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.printed {
		fmt.Fprint(s.writer, "\033[3A\n") // move cursor before table
	}

	for _, h := range s.headers {
		fmt.Fprintf(s.writer, "%s\t", h)
	}
	fmt.Fprintln(s.writer)

	for _, h := range s.headers {
		progress := s.progress[h]
		switch {
		case progress >= 0:
			fmt.Fprintf(s.writer, "%6.2f%%\t", progress)
		case progress == progressNotAvailable:
			fmt.Fprintf(s.writer, "N/A\t")
		case progress == downloadFailure:
			fmt.Fprintf(s.writer, "Error\t")
		default:
			fmt.Fprintf(s.writer, "Error: %.0f\t", progress)
		}
	}
	fmt.Fprintln(s.writer)
	s.printed = true

	return s.writer.Flush()
}
