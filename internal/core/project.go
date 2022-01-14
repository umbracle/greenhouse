package core

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/greenhouse/internal/solidity"
	"github.com/umbracle/greenhouse/internal/standard"
)

type Project struct {
	logger hclog.Logger
	config *Config

	// structure of the contracts directory
	// dir *Directory

	// svm is the solidity version manager
	svm *solidity.Solidity

	// list of compiled sources/outputs
	//sources []*Source

	// metadata holds the structure of sources and contracts
	metadata2 *Metadata

	//metadata map[string]*Source

	// map of the source code name to the location
	remappings map[string]string

	// directory where the remaps are
	remapDir string
}

func NewProject(logger hclog.Logger, config *Config) (*Project, error) {
	p := &Project{
		logger:     logger,
		config:     config,
		remappings: map[string]string{},
	}

	/*
		if err := p.initDirectory(); err != nil {
			return nil, err
		}
	*/
	if err := p.initSources(); err != nil {
		return nil, err
	}

	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	p.svm = solidity.NewSolidity(filepath.Join(dirname, ".greenhouse"))

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

	// eventually we want to merge this dir with the one for solidity compiler
	// create some directory for standard remappings
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	remapDir := filepath.Join(dirname, ".greenhouse/standard/greenhouse")
	if err := os.MkdirAll(remapDir, os.ModePerm); err != nil {
		return err
	}
	p.remapDir = filepath.Join(dirname, ".greenhouse/standard")
	for name, raw := range standard.SystemContracts {
		filename := filepath.Base(name)
		fullPath := filepath.Join(remapDir, filename)

		if err := ioutil.WriteFile(fullPath, []byte(raw), 0755); err != nil {
			panic(err)
		}
		p.remappings[name] = fullPath
	}

	/*
		fmt.Println("XX")

		for _, f := range p.dir.Files {
			fmt.Println("-- f --")
			fmt.Println(f)

			src := &Source{
				ModTime: f.Modtime,
				Hash:    hash(f.Content),
				Path:    f.Path,
			}
		}

		panic("X")
	*/

	/*
		// read the contracts in output
		files, err := ioutil.ReadDir(filepath.Join(".greenhouse", "contracts"))
		if err != nil {
			return err
		}

		sources := []*Source{}
		for _, file := range files {
			//var source Source
			data, err := ioutil.ReadFile(filepath.Join(".greenhouse", "contracts", file.Name()))
			if err != nil {
				return err
			}
			var source *Source
			if err := json.Unmarshal(data, &source); err != nil {
				return err
			}
			sources = append(sources, source)
		}
		p.sources = sources
	*/

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

type Metadata struct {
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

func (m *Metadata) Diff(files []*File1) ([]*FileDiff, error) {
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
