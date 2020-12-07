package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgressNotifierWrite(t *testing.T) {
	bytes := []byte("12345")

	ch := make(chan float64, 1)
	pn := &progressNotifier{
		total:      2 * int64(len(bytes)),
		progressCh: ch,
	}

	n, err := pn.Write(bytes)

	assert.NoError(t, err)
	assert.Equal(t, len(bytes), n, "Invalid n")
	assert.Equal(t, 50.0, <-ch, "Invalid progress")

	n, err = pn.Write(bytes)

	assert.NoError(t, err)
	assert.Equal(t, len(bytes), n, "Invalid n")
	assert.Equal(t, 100.0, <-ch, "Invalid progress")
}

func TestDownloadFileWithProgress(t *testing.T) {
	fileURL := "https://github.com/tenntenn/gopher-stickers/raw/master/png/angry.png"
	filePath := filepath.Join(t.TempDir(), "angry.png")

	output := downloadFile(filePath, fileURL)

	var progress float64
	for progress = range output {
		assert.Greater(t, progress, 0.0, "Unexpected progress")
	}
	assert.Equal(t, 100.0, progress, "Download is not completed")
	assert.FileExists(t, filePath)
}

func TestDownloadFileWithoutProgress(t *testing.T) {
	fileURL := "https://github.com"
	filePath := filepath.Join(t.TempDir(), "github.com")

	output := downloadFile(filePath, fileURL)

	progress := <-output
	assert.Equal(t, ProgressNotAvailable, progress, "Unexpected progress")
	assert.FileExists(t, filePath)
}
