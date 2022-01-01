package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/umbracle/greenhouse/internal/lib/dag"
	soltestlib "github.com/umbracle/greenhouse/internal/sol-test-lib"
)

type Output struct {
	Contracts map[string]*Artifact
	Sources   map[string]*SourceF
	Version   string
}

type SourceF struct {
	AST *ASTNode
}

type Artifact struct {
	Abi           string `json:"abi"`
	Bin           string `json:"bin"`
	BinRuntime    string `json:"bin-runtime"`
	SrcMap        string `json:"srcmap"`
	SrcMapRuntime string `json:"srcmap-runtime"`
}

func (p *Project) loadDependencies() error {
	fmt.Println(p.config.Dependencies)

	for name, tag := range p.config.Dependencies {
		args := []string{
			"install",
			name + "@" + tag,
			"--prefix",
			".greenhouse",
		}
		cmd := exec.Command("npm", args...)
		log.Printf("Running command and waiting for it to finish...")
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	}
	return nil
}

// Compile compiles the application
func (p *Project) Compile() error {

	cc := [][]string{
		{"greenhouse", "testify.sol", soltestlib.GetSolcTest()},
		{"greenhouse", "console.sol", soltestlib.GetConsoleLib()},
	}
	for _, c := range cc {
		// add the ds-test contract
		sourcePath := filepath.Join(".greenhouse", "system", c[0])
		if err := os.MkdirAll(sourcePath, 0755); err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(filepath.Join(sourcePath, c[1]), []byte(c[2]), 0644); err != nil {
			panic(err)
		}
	}

	if err := p.loadDependencies(); err != nil {
		return err
	}

	dirMap := map[string]*File{}
	for _, f := range p.dir.Files {
		dirMap[f.Path] = f
	}

	// detect dependencies
	for _, f := range p.dir.Files {
		f.DependsOn = parseDependencies(f.Content)
		f.Pragma = parsePragma(f.Content)
	}

	// build dag map
	dd := &dag.Dag{}
	for _, f := range p.dir.Files {
		dd.AddVertex(f)
	}
	// add edges
	for _, src := range p.dir.Files {
		for _, dst := range src.DependsOn {
			dd.AddEdge(dag.Edge{
				Src: src,
				Dst: dirMap[dst],
			})
		}
	}

	for _, f := range p.dir.Files {
		fmt.Println("---")
		fmt.Println(f.Path)
		fmt.Println(f.DependsOn)
	}

	fmt.Println("-- find components --")

	// find components in the dag
	components := findComponents(dirMap, dd)

	fmt.Println(components)

	sources := map[string]*Source{}
	outputs := []*Output{}
	for _, comp := range components {

		pathFiles := []*File{}
		paths := []string{}
		for _, p := range comp {
			pathFiles = append(pathFiles, dirMap[p])
			paths = append(paths, dirMap[p].RealPath)
		}

		// get version for this..
		var versionPragma []string
		for _, ff := range pathFiles {
			for _, i := range strings.Split(ff.Pragma[0], " ") {
				if !contains(versionPragma, i) {
					versionPragma = append(versionPragma, i)
				}
			}
		}

		fmt.Println("-- versionPragma --")
		fmt.Println(versionPragma)

		solidityVersion, err := version.NewVersion(p.config.Solidity)
		if err != nil {
			panic(err)
		}
		versionConstraint, err := version.NewConstraint(strings.Join(versionPragma, ", "))
		if err != nil {
			panic(err)
		}

		fmt.Println("---")
		fmt.Println(solidityVersion)
		fmt.Println(versionConstraint)

		if !versionConstraint.Check(solidityVersion) {
			panic("not match")
		}

		remapping := map[string]string{
			"@openzeppelin/contracts": "/home/ferran/go/src/github.com/umbracle/greenhouse/.greenhouse/node_modules/@openzeppelin/contracts",
			"greenhouse":              "/home/ferran/go/src/github.com/umbracle/greenhouse/.greenhouse/system/greenhouse",
		}
		output, err := p.compileImpl(solidityVersion.String(), paths, remapping)
		if err != nil {
			return err
		}

		outputs = append(outputs, output)

		for name, x := range output.Sources {
			if _, ok := sources[name]; ok {
				continue
			}

			contractNames := []string{}
			x.AST.Visit(ASTContractDefinition, func(n *ASTNode) {
				contractNames = append(contractNames, n.GetAttribute("name").(string))
			})

			artifacts := []*Contract{}
			for _, cc := range contractNames {
				res, ok := output.Contracts[name+":"+cc]
				if !ok {
					panic("not found?")
				}
				artifacts = append(artifacts, &Contract{
					Name:     cc,
					Artifact: res,
				})
			}
			sources[name] = &Source{
				Path:      name,
				ModTime:   time.Now(),
				AST:       x.AST,
				Contract:  artifacts,
				Artifacts: contractNames,
				Imports:   paths,
			}
		}
	}

	fmt.Println("-- outputs --")
	fmt.Println(outputs)

	{
		xx := map[string]*Source1{}
		for k, v := range sources {
			fmt.Println("ZZZZZZZ")
			fmt.Println(v.Artifacts, v.X().Artifacts)
			xx[k] = v.X()
		}
		data, err := json.Marshal(xx)
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(filepath.Join(".greenhouse", "metadata.json"), data, 0644); err != nil {
			panic(err)
		}
	}

	for name, src := range sources {
		sourcePath := filepath.Join(".greenhouse", "contracts", name)
		if err := os.MkdirAll(sourcePath, 0755); err != nil {
			panic(err)
		}
		for _, cc := range src.Contract {
			data, err := json.Marshal(cc.Artifact)
			if err != nil {
				panic(err)
			}
			if err := ioutil.WriteFile(filepath.Join(sourcePath, cc.Name+".json"), data, 0644); err != nil {
				panic(err)
			}
		}
	}

	p.metadata = sources
	return nil
}

func (p *Project) compileImpl(version string, files []string, remapping map[string]string) (*Output, error) {
	if !p.svm.Exists(version) {
		fmt.Printf("Downloading compiler: %s\n", version)

		if err := p.svm.Download(version); err != nil {
			panic(err)
		}
	}

	path := p.svm.Path(version)

	args := []string{
		"--combined-json",
		"bin,bin-runtime,srcmap-runtime,abi,srcmap,ast",
	}
	if len(remapping) != 0 {
		for k, v := range remapping {
			args = append(args, k+"="+v)
		}
	}
	if len(files) != 0 {
		args = append(args, files...)
	}

	fmt.Println("-- args --")
	fmt.Println(args)

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(path, args...)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to compile: %s", stderr.String())
	}

	fmt.Println(stdout.String())

	var output *Output
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, err
	}
	return output, nil
}

func findComponents(dirMap map[string]*File, d *dag.Dag) [][]string {
	checked := map[string]struct{}{}

	var walk func(f *File) []string
	walk = func(f *File) (res []string) {
		if _, ok := checked[f.Path]; ok {
			return
		}

		checked[f.Path] = struct{}{}

		res = append(res, f.Path)
		for _, v := range d.GetOutbound(f) {
			if v.(*File) != nil {
				res = append(res, walk(v.(*File))...)
			}
		}
		for _, v := range d.GetInbound(f) {
			if v.(*File) != nil {
				res = append(res, walk(v.(*File))...)
			}
		}

		return
	}

	components := [][]string{}

	for name, m := range dirMap {
		//fmt.Println("----- xx ------")
		//fmt.Println(name)
		//fmt.Println(checked)

		// for this file find all the dependencies
		if _, ok := checked[name]; ok {
			continue
		}

		// for this file find all the components connected
		res := walk(m)

		components = append(components, res)

	}

	//fmt.Println("-- components --")
	//fmt.Println(components)

	return components
}

var (
	importRegexp = regexp.MustCompile(`import "(.*)"`)
	pragmaRegexp = regexp.MustCompile(`pragma\s+solidity\s+(.*);`)
)

func parseDependencies(contract string) []string {
	res := importRegexp.FindStringSubmatch(contract)
	if len(res) == 0 {
		return []string{}
	}
	return res[1:]
}

func parsePragma(contract string) []string {
	res := pragmaRegexp.FindStringSubmatch(contract)
	if len(res) == 0 {
		panic("pragma not found")
	}
	return res[1:]
}

type ASTType string

var (
	ASTContractDefinition ASTType = "ContractDefinition"
)

type ASTNode struct {
	Children   []*ASTNode             `json:"children"`
	Name       ASTType                `json:"name"`
	Src        string                 `json:"src"`
	Id         int                    `json:"id"`
	Attributes map[string]interface{} `json:"attributes"`
}

func (a *ASTNode) GetAttributeOk(name string) (interface{}, bool) {
	k, v := a.Attributes[name]
	return k, v
}

func (a *ASTNode) GetAttribute(name string) interface{} {
	return a.Attributes[name]
}

func (a *ASTNode) Visit(name ASTType, handler func(n *ASTNode)) {
	if a.Name == name {
		handler(a)
	}
	for _, child := range a.Children {
		child.Visit(name, handler)
	}
}

func contains(s []string, i string) bool {
	for _, j := range s {
		if j == i {
			return true
		}
	}
	return false
}
