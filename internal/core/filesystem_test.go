package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSystem(t *testing.T) {
	files, err := Walk("./fixtures/filesystem")
	assert.NoError(t, err)

	names := []string{}
	for _, file := range files {
		names = append(names, file.Path)
	}

	expect := []string{
		"fixtures/filesystem/a.txt",
		"fixtures/filesystem/b.txt",
		"fixtures/filesystem/c/d.txt",
	}
	assert.Equal(t, names, expect)
}
