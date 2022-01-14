package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/umbracle/greenhouse/internal/lib/dag"
	"github.com/umbracle/greenhouse/internal/solidity"
)

func (p *Project) loadMetadata() ([]*FileDiff, error) {
	// this is something you always have to load
	files, err := Walk(p.config.Contracts)
	if err != nil {
		panic(err)
	}

	var metadata *Metadata

	metadataPath := filepath.Join(".greenhouse", "metadata.json")
	exists, err := existsFile(metadataPath)
	if err != nil {
		return nil, err
	}
	if exists {
		// load the metadata from a file
		data, err := ioutil.ReadFile(metadataPath)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(data, &metadata); err != nil {
			panic(err)
		}
	} else {
		metadata = &Metadata{
			Sources: map[string]*Source{},
			Output:  map[string]*SolcOutput{},
		}
	}

	diffFiles, err := metadata.Diff(files)
	if err != nil {
		panic(err)
	}

	p.metadata2 = metadata
	return diffFiles, nil
}

// Compile compiles the application
func (p *Project) Compile() error {
	diffFiles, err := p.loadMetadata()
	if err != nil {
		panic(err)
	}
	if len(diffFiles) == 0 {
		// nothing to compile
		return nil
	}

	/*
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
	*/

	/*
		if err := p.loadDependencies(); err != nil {
			return err
		}
	*/

	// source files
	sources2 := p.metadata2.Sources

	fmt.Println("-- diff --")
	fmt.Println(diffFiles)

	remappings := map[string]string{}

	// detect dependencies / ONLY DO THIS FOR NEW/MOD FILES
	diffSources := []*Source{}
	for _, f := range diffFiles {
		if f.Type == FileDiffDel {
			continue
		}
		content, err := ioutil.ReadFile(f.Path)
		if err != nil {
			panic(err)
		}

		// only process add and delete files. Check again
		// imports, pragma and dependencies.
		// For now we update the full source list files?
		// Consider that we are overriding the other!
		imports := parseDependencies(string(content))
		cleanImports := []string{}

		fmt.Println("--- file ", f.Path)
		fmt.Println(imports)

		// this is the parent path of this file, we have to resolve all the import
		// files with respect to this parent
		parentPath := filepath.Dir(f.Path)

		// clean the imports so that we can convert
		// ./d.sol to ./contracts/d.sol wrt the contracts repo
		for _, im := range imports {
			// local
			if !strings.HasPrefix(im, ".") {
				fmt.Println(im)
				if fullPath, ok := p.remappings[im]; ok {
					remappings[im] = fullPath
				} else {
					// imported file
					panic("UNIVERSTAL PATH TODO")
				}
			} else {
				cleanImports = append(cleanImports, filepath.Join(parentPath, im))
			}
		}

		src := &Source{
			Imports: cleanImports,
			Version: parsePragma(string(content)),
			ModTime: f.Mod,
			Path:    f.Path,
			Hash:    hash(string(content)),
		}
		sources2[f.Path] = src
		diffSources = append(diffSources, src)
	}

	fmt.Println("_ LIST SOURCES _")
	for _, f := range sources2 {
		fmt.Println("---")
		fmt.Println(f.Path)
		fmt.Println(f.Imports)
	}

	// build dag map (move this to own repo)
	dd := &dag.Dag{}
	for _, f := range sources2 {
		dd.AddVertex(f)
	}
	// add edges
	for _, src := range sources2 {
		for _, dst := range src.Imports {
			dst, ok := sources2[dst]
			if !ok {
				fmt.Println(dst)
				panic("BUG")
			}
			dd.AddEdge(dag.Edge{
				Src: src,
				Dst: dst,
			})
		}
	}

	rawComponents := dd.FindComponents()

	components := [][]string{}
	for _, comp := range rawComponents {
		// if any of the components of component are in diffSources
		// we have to recompile this component
		found := false
		for _, i := range comp {
			for _, j := range diffSources {
				if i == j {
					found = true
				}
			}
		}
		if found {
			subComp := []string{}
			for _, i := range comp {
				subComp = append(subComp, i.(*Source).Path)
			}
			components = append(components, subComp)
		}
	}

	fmt.Println("-- real components --")
	fmt.Println(components)

	// solidity version we use to compile
	solidityVersion, err := version.NewVersion(p.config.Solidity)
	if err != nil {
		panic(err)
	}

	contracts := map[string]*solidity.Artifact{}
	//outputs := map[string]*SolcOutput{}

	// generate the outputs and compile
	for _, comp := range components {
		fmt.Println("-- component --")
		fmt.Println(comp)

		pragmas := []string{}
		for _, i := range comp {
			pragmas = append(pragmas, strings.Split(sources2[i].Version[0], " ")...)
		}
		pragmas = unique(pragmas)

		versionConstraint, err := version.NewConstraint(strings.Join(pragmas, ", "))
		if err != nil {
			panic(err)
		}
		if !versionConstraint.Check(solidityVersion) {
			panic("not match in solidity compiler")
		}

		// compile
		input := &solidity.Input{
			Version:    solidityVersion.String(),
			Files:      comp,
			Remappings: remappings,
		}
		output, err := p.svm.Compile(input)
		if err != nil {
			return err
		}

		id := uuid.New()
		out := &SolcOutput{
			Id:     id.String(),
			Output: output,
		}
		p.metadata2.Output[id.String()] = out

		for _, i := range comp {
			sources2[i].BuildInfo = id.String()
		}

		// what else do we do here?
		for name, c := range output.Contracts {
			contracts[name] = c
		}
	}

	{
		// write artifacts!
		for name, contract := range contracts {
			// name has the format <path>:<contract>
			// remove the contract name
			spl := strings.Split(name, ":")
			path, name := spl[0], spl[1]

			fmt.Println("--xx")
			fmt.Println(path)

			if strings.HasPrefix(path, p.remapDir) {
				// this comes from remap its an address of type '/home/.../greenhouse/Console.sol'
				// we have to remove that mapping
				path = strings.TrimPrefix(path, p.remapDir)
			}
			// remove the contracts path in the destination name
			sourcePath := filepath.Join(".greenhouse", path)
			if err := os.MkdirAll(sourcePath, 0755); err != nil {
				panic(err)
			}

			// write the file
			raw, err := json.Marshal(contract)
			if err != nil {
				panic(err)
			}
			if err := ioutil.WriteFile(filepath.Join(sourcePath, name+".json"), raw, 0644); err != nil {
				panic(err)
			}
		}
	}

	{
		// write metadata!

		// clean any extra output that is not references anymore
		deleteNames := []string{}
		for name := range p.metadata2.Output {
			found := false
			for _, src := range p.metadata2.Sources {
				if src.BuildInfo == name {
					found = true
					break
				}
			}
			if !found {
				deleteNames = append(deleteNames, name)
			}
		}
		for _, name := range deleteNames {
			delete(p.metadata2.Output, name)
		}

		raw, err := json.Marshal(p.metadata2)
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(filepath.Join(".greenhouse", "metadata.json"), raw, 0644); err != nil {
			panic(err)
		}
	}

	return nil
}

var (
	importRegexp = regexp.MustCompile(`import (".*"|'.*')`)
	pragmaRegexp = regexp.MustCompile(`pragma\s+solidity\s+(.*);`)
)

func parseDependencies(contract string) []string {
	res := importRegexp.FindAllStringSubmatch(contract, -1)
	if len(res) == 0 {
		return []string{}
	}

	clean := []string{}
	for _, j := range res {
		i := j[1]
		i = strings.Trim(i, "'")
		i = strings.Trim(i, "\"")
		clean = append(clean, i)
	}
	return clean
}

func parsePragma(contract string) []string {
	res := pragmaRegexp.FindStringSubmatch(contract)
	if len(res) == 0 {
		panic("pragma not found")
	}
	return res[1:]
}

func unique(a []string) []string {
	b := []string{}
	for _, i := range a {
		found := false
		for _, j := range b {
			if i == j {
				found = true
			}
		}
		if !found {
			b = append(b, i)
		}
	}
	return b
}
