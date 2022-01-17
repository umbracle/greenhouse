package core

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/greenhouse/internal/solidity"
)

type Project struct {
	logger hclog.Logger
	config *Config

	// wrapper to compile solidity
	sol *solidity.Solidity

	// state holds the structure of sources and contracts
	state *State
}

func NewProject(logger hclog.Logger, config *Config) (*Project, error) {
	p := &Project{
		logger: logger,
		config: config,
	}
	if err := p.initSources(); err != nil {
		return nil, err
	}

	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	p.sol = solidity.NewSolidity(filepath.Join(dirname, ".greenhouse"))
	return p, nil
}

func (p *Project) initSources() error {
	// create the config dir if does not exists
	if err := os.MkdirAll(".greenhouse", os.ModePerm); err != nil {
		return err
	}
	dirs := []string{"contracts"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(".greenhouse", dir), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// minified version
type Source struct {
	ModTime   time.Time
	BuildInfo string
	Path      string
	Hash      string
	Imports   []string
	Version   []string
	Artifacts []string
}

type Contract struct {
	Name     string
	Artifact *solidity.Artifact
}

type State struct {
	Sources map[string]*Source
	Output  map[string]*SolcOutput
}

type FileDiffType string

const (
	FileDiffAdd FileDiffType = "add"
	FileDiffDel FileDiffType = "del"
	FileDiffMod FileDiffType = "mod"
)

type FileDiff struct {
	Path string
	Type FileDiffType
	Mod  time.Time
}

func (m *State) Diff(files []*File1) ([]*FileDiff, error) {
	diff := []*FileDiff{}

	visited := map[string]struct{}{}
	for _, file := range files {
		visited[file.Path] = struct{}{}

		if src, ok := m.Sources[file.Path]; ok {
			if !src.ModTime.Equal(file.ModTime) {
				// mod file
				diff = append(diff, &FileDiff{
					Path: file.Path,
					Type: FileDiffMod,
					Mod:  file.ModTime,
				})
			}
		} else {
			// new file
			diff = append(diff, &FileDiff{
				Path: file.Path,
				Type: FileDiffAdd,
				Mod:  file.ModTime,
			})
		}
	}

	for _, src := range m.Sources {
		if _, ok := visited[src.Path]; !ok {
			// deleted
			diff = append(diff, &FileDiff{
				Path: src.Path,
				Type: FileDiffDel,
				Mod:  time.Time{},
			})
		}
	}

	return diff, nil
}

func existsFile(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

type SolcOutput struct {
	Id     string
	Output *solidity.Output
}
