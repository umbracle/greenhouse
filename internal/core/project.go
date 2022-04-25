package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/greenhouse/internal/solidity"
	"github.com/umbracle/greenhouse/internal/standard"
	"github.com/umbracle/greenhouse/internal/state"
)

type Project struct {
	logger hclog.Logger
	config *Config

	// wrapper to compile solidity
	sol *solidity.Solidity

	// state holds the structure of sources and contracts
	state *state.State

	// list of standard remapping contracts
	remappings map[string]string

	// path for the imported lib directory
	libDirectory string
}

func NewProject(logger hclog.Logger, config *Config) (*Project, error) {
	p := &Project{
		logger:     logger,
		config:     config,
		remappings: map[string]string{},
	}
	if err := p.initSources(); err != nil {
		return nil, err
	}

	state, err := state.NewState()
	if err != nil {
		return nil, err
	}
	p.state = state

	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}
	dirname = filepath.Join(dirname, ".greenhouse")

	p.sol = solidity.NewSolidity(dirname)

	// write the standard contracts to system folder
	libDir := filepath.Join(dirname, "lib")
	for c, code := range standard.SystemContracts {
		stanLib := filepath.Join(libDir, c)
		if err := os.MkdirAll(filepath.Dir(stanLib), 0700); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(stanLib, []byte(code), 0755); err != nil {
			return nil, err
		}
		p.remappings[c] = stanLib
	}
	p.libDirectory = libDir

	if err := p.loadMetadata(); err != nil {
		return nil, err
	}
	// right after the start figure out if there are any tainted nodes
	if err := p.findLocalDiff(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Project) initSources() error {
	// create the config dir if does not exists
	if err := os.MkdirAll(".greenhouse", os.ModePerm); err != nil {
		return err
	}
	dirs := []string{"contracts"}
	for _, dir := range dirs {
		if err := os.MkdirAll(p.getFile(".greenhouse", dir), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

type FileDiffType string

const (
	FileDiffAdd FileDiffType = "add"
	FileDiffDel FileDiffType = "del"
	FileDiffMod FileDiffType = "mod"
)

type FileDiff struct {
	Path   string
	Type   FileDiffType
	Mod    time.Time
	Source *state.Source
}

func Diff(sources []*state.Source, files []*File1) ([]*FileDiff, error) {
	diff := []*FileDiff{}

	sourcesMap := map[string]*state.Source{}
	for _, src := range sources {
		sourcesMap[filepath.Join(src.Dir, src.Filename)] = src
	}

	readAndParseFile := func(path string) (*state.Source, error) {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		source, err := parseSource(string(content), path)
		if err != nil {
			return nil, err
		}
		source.Tainted = true
		return source, nil
	}

	visited := map[string]struct{}{}
	for _, file := range files {
		visited[file.Path] = struct{}{}

		if src, ok := sourcesMap[file.Path]; ok {
			if !src.ModTime.Equal(file.ModTime) {
				// mod file
				source, err := readAndParseFile(file.Path)
				if err != nil {
					return nil, err
				}
				diff = append(diff, &FileDiff{
					Path:   file.Path,
					Type:   FileDiffMod,
					Mod:    file.ModTime,
					Source: source,
				})
			}
		} else {
			// new file
			source, err := readAndParseFile(file.Path)
			if err != nil {
				return nil, err
			}

			diff = append(diff, &FileDiff{
				Path:   file.Path,
				Type:   FileDiffAdd,
				Mod:    file.ModTime,
				Source: source,
			})
		}
	}

	for path := range sourcesMap {
		if _, ok := visited[path]; !ok {
			// deleted
			diff = append(diff, &FileDiff{
				Path: path,
				Type: FileDiffDel,
				Mod:  time.Time{},
			})
		}
	}

	return diff, nil
}

func (p *Project) getFile(path ...string) string {
	fullPath := append([]string{}, path...)
	if p.config.DataDir != "" {
		fullPath = append([]string{p.config.DataDir}, fullPath...)
	}
	return filepath.Join(fullPath...)
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

func (p *Project) loadMetadata() error {
	var metadata *metadataFormat

	metadataPath := p.getFile(".greenhouse", "metadata.json")
	exists, err := existsFile(metadataPath)
	if err != nil {
		return err
	}
	if exists {
		// load the metadata from a file
		data, err := ioutil.ReadFile(metadataPath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &metadata); err != nil {
			return err
		}
	} else {
		metadata = &metadataFormat{
			Sources:   []*state.Source{},
			Contracts: []*state.Contract{},
		}
	}

	// fill in the state with the metadata object
	for _, src := range metadata.Sources {
		if err := p.state.UpsertSource(src); err != nil {
			return err
		}
	}

	// create the contracts
	for _, contract := range metadata.Contracts {
		if err := p.state.UpsertContract(contract); err != nil {
			return err
		}
	}
	return nil
}

func parseSource(content, path string) (*state.Source, error) {
	// new file
	dir, filename := filepath.Dir(path), filepath.Base(path)

	file, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	oldImports := parseDependencies(string(content))
	parentPath := filepath.Dir(path)

	// clean the imports so that we can convert relative file names to
	// their path with respect to the contracts repo (i.e ./d.sol to ./contracts/d.sol)
	cleanImports := []string{}
	for _, im := range oldImports {
		// local
		if !strings.HasPrefix(im, ".") {
			cleanImports = append(cleanImports, im)
		} else {
			cleanImports = append(cleanImports, filepath.Join(parentPath, im))
		}
	}

	pragma, err := parsePragma(string(content))
	if err != nil {
		return nil, err
	}

	source := &state.Source{
		Dir:      dir,
		Filename: filename,
		Version:  pragma,
		Imports:  cleanImports,
		ModTime:  file.ModTime(),
	}
	return source, nil
}
