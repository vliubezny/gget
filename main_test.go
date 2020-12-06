package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateURL(t *testing.T) {
	cases := []struct {
		fileURL string
		isValid bool
	}{
		{"https://github.com/tenntenn/gopher-stickers/raw/master/png/angry.png", true},
		{"https://github.com/angry.png", true},
		{"https://github.com/", false},
		{"https://github.com", false},
		{"github.com", false},
		{"/", false},
		{"", false},
	}

	for _, c := range cases {
		t.Run(c.fileURL, func(t *testing.T) {
			err := validateURL(c.fileURL)
			if c.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
