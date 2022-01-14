package core

import (
	"os"
	"path/filepath"
	"time"
)

type File1 struct {
	Path    string
	ModTime time.Time
}

func Walk(dirPath string) ([]*File1, error) {
	files := []*File1{}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		files = append(files, &File1{
			Path:    path,
			ModTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
