package main

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgressNotifierWrite(t *testing.T) {
	bytes := []byte("12345")
	bytesLen := int64(len(bytes))
	var cases = []struct {
		total            int64
		expectedProgress float64
	}{
		{-1, ProgressNotAvailable},
		{0, ProgressNotAvailable},
		{2 * bytesLen, 50.0},
		{bytesLen, 100.0},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("Total: %d", c.total), func(t *testing.T) {
			ch := make(chan float64, 1)
			pn := &progressNotifier{
				total:      c.total,
				progressCh: ch,
			}

			n, err := pn.Write(bytes)

			assert.NoError(t, err)
			assert.Equalf(t, len(bytes), n, "Invalid n")
			assert.Equalf(t, c.expectedProgress, <-ch, "Invalid progress")
		})
	}
}

func TestDownloadFile(t *testing.T) {
	fileURL := "https://github.com/tenntenn/gopher-stickers/raw/master/png/angry.png"
	filePath := filepath.Join(t.TempDir(), "angry.png")

	output := downloadFile(filePath, fileURL)

	var progress float64
	for progress = range output {
		assert.Greaterf(t, progress, 0.0, "Unexpected progress: %f", progress)
	}
	assert.Equal(t, 100.0, progress, "Download is not completed")
	assert.FileExists(t, filePath)
}
