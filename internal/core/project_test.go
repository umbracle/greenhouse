package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProject_FileSystemDiff(t *testing.T) {
	// detect differences between a metadata file and a filesystem

	time0 := time.Unix(0, 10)
	time1 := time.Unix(0, 20)

	m := Metadata{
		Sources: map[string]*Source{
			"a.txt": {
				Path:    "a.txt",
				ModTime: time0,
			},
			"b.txt": {
				Path:    "b.txt",
				ModTime: time0,
			},
			"c.txt": {
				Path:    "c.txt",
				ModTime: time0,
			},
		},
	}

	files := []*File1{
		{"a.txt", time0}, // not modified
		{"b.txt", time1}, // modified
		{"d.txt", time1}, // new
	}

	expectedDiff := []*FileDiff{
		{"b.txt", FileDiffMod, time1},
		{"d.txt", FileDiffAdd, time1},
		{"c.txt", FileDiffDel, time.Time{}},
	}

	res, err := m.Diff(files)
	assert.NoError(t, err)
	assert.Equal(t, res, expectedDiff)
}
