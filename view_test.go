package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

const expectedTable = `file1     file2     file3     file4     file5      
  0.00%   100.00%   N/A       Error     Error: -5  
`

func TestProgressStatePrint(t *testing.T) {
	buf := new(bytes.Buffer)
	s := newProgressState([]string{"file1", "file2", "file3", "file4", "file5"}, buf)
	s.update("file1", 0.0)
	s.update("file2", 100.0)
	s.update("file3", progressNotAvailable)
	s.update("file4", downloadFailure)
	s.update("file5", -5.0)

	err := s.print()

	assert.NoError(t, err)
	assert.Equal(t, expectedTable, buf.String())
}
