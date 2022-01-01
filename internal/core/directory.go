package core

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Directory struct {
	Files []*File `json:"files"`
}

type File struct {
	Path      string
	RealPath  string
	Content   string
	Modtime   time.Time
	DependsOn []string
	Pragma    []string
}

type FileUpdate struct {
	Name string
}

func Walk(dirPath string) (*Directory, error) {
	dir := &Directory{}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relativePath := "." + strings.TrimPrefix(path, dirPath)

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		dir.Files = append(dir.Files, &File{
			Path:     relativePath,
			RealPath: path,
			Content:  string(content),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dir, nil
}
