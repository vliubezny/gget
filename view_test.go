package main

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const expectedTable = `file1     file2     file3     file4     file5      
  0.00%   100.00%   N/A       Error     Error: -5  
`

func TestPrintTable(t *testing.T) {
	buf := new(bytes.Buffer)
	initWriter(buf)
	statuses := []float64{0.0, 100.0, ProgressNotAvailable, DownloadFailure, -5.0}
	headers := make([]string, len(statuses))
	state := &sync.Map{}
	for i, status := range statuses {
		filePath := fmt.Sprintf("file%d", i+1)
		headers[i] = filePath
		state.Store(filePath, status)
	}

	printTable(headers, state)

	assert.Equal(t, expectedTable, buf.String())
}
