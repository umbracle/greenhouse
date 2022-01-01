package core

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/umbracle/greenhouse/internal/solidity"
)

type Project struct {
	logger hclog.Logger
	config *Config

	// structure of the contracts directory
	dir *Directory

	// svm is the solidity version manager
	svm *solidity.SolidityVersionManager

	// list of compiled sources/outputs
	sources []*Source

	metadata map[string]*Source
}

func NewProject(logger hclog.Logger, config *Config) (*Project, error) {
	p := &Project{
		logger: logger,
		config: config,
	}

	if err := p.initDirectory(); err != nil {
		return nil, err
	}
	if err := p.initSources(); err != nil {
		return nil, err
	}

	s, err := solidity.NewSolidityVersionManager("")
	if err != nil {
		return nil, err
	}
	p.svm = s

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

func (p *Project) initDirectory() error {
	dir := &Directory{
		Files: []*File{},
	}

	err := filepath.Walk(p.config.Contracts, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relativePath := "." + strings.TrimPrefix(path, p.config.Contracts)

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
		return err
	}
	p.dir = dir

	return nil
}

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Source is a source file in the contracts folder
type Source struct {
	ModTime   time.Time
	Hash      string
	Path      string
	Imports   []string
	Version   []string
	Artifacts []string
	AST       *ASTNode

	// extra
	Contract []*Contract
}

func (s *Source) X() *Source1 {
	return &Source1{
		ModTime:   s.ModTime,
		Hash:      s.Hash,
		Imports:   s.Imports,
		Version:   s.Version,
		Artifacts: s.Artifacts,
	}
}

// minified version
type Source1 struct {
	ModTime   time.Time
	Hash      string
	Imports   []string
	Version   []string
	Artifacts []string
}

type Contract struct {
	Name     string
	Artifact *Artifact
}
