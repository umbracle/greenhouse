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
	"github.com/umbracle/greenhouse/internal/dag"
	"github.com/umbracle/greenhouse/internal/solidity"
)

func (p *Project) loadMetadata() ([]*FileDiff, error) {
	// this is something you always have to load
	files, err := Walk(p.config.Contracts)
	if err != nil {
		return nil, err
	}

	var metadata *State

	metadataPath := filepath.Join(".greenhouse", "metadata.json")
	exists, err := existsFile(metadataPath)
	if err != nil {
		return nil, err
	}
	if exists {
		// load the metadata from a file
		data, err := ioutil.ReadFile(metadataPath)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &metadata); err != nil {
			return nil, err
		}
	} else {
		metadata = &State{
			Sources: map[string]*Source{},
			Output:  map[string]*SolcOutput{},
		}
	}

	diffFiles, err := metadata.Diff(files)
	if err != nil {
		return nil, err
	}

	p.state = metadata
	return diffFiles, nil
}

// Compile compiles the application
func (p *Project) Compile() error {
	// solidity version we use to compile
	solidityVersion, err := version.NewVersion(p.config.Solidity)
	if err != nil {
		return err
	}

	diffFiles, err := p.loadMetadata()
	if err != nil {
		return nil
	}
	if len(diffFiles) == 0 {
		// nothing to compile
		return nil
	}

	sources := p.state.Sources

	remappings := map[string]string{}

	// detect dependencies only for new and modified files
	diffSources := []*Source{}
	for _, f := range diffFiles {
		if f.Type == FileDiffDel {
			continue
		}
		content, err := ioutil.ReadFile(f.Path)
		if err != nil {
			return err
		}

		imports := parseDependencies(string(content))
		parentPath := filepath.Dir(f.Path)

		// clean the imports so that we can convert relative file names to
		// their path with respect to the contracts repo (i.e ./d.sol to ./contracts/d.sol)
		cleanImports := []string{}
		for _, im := range imports {
			// local
			if !strings.HasPrefix(im, ".") {
				// absolute path, check if we can resolve it using the remappings
				if fullPath, exists := p.remappings[im]; exists {
					remappings[im] = fullPath
				} else {
					return fmt.Errorf("absolute paths cannot be resolved yet")
				}
			} else {
				cleanImports = append(cleanImports, filepath.Join(parentPath, im))
			}
		}

		pragma, err := parsePragma(string(content))
		if err != nil {
			return err
		}
		src := &Source{
			Imports: cleanImports,
			Version: pragma,
			ModTime: f.Mod,
			Path:    f.Path,
			Hash:    hash(string(content)),
		}
		sources[f.Path] = src
		diffSources = append(diffSources, src)
	}

	// build dag map (move this to own repo)
	dd := &dag.Dag{}
	for _, f := range sources {
		dd.AddVertex(f)
	}
	// add edges
	for _, src := range sources {
		for _, dst := range src.Imports {
			dst, ok := sources[dst]
			if !ok {
				panic(fmt.Errorf("BUG: elem in DAG not found: %s", dst))
			}
			dd.AddEdge(dag.Edge{
				Src: src,
				Dst: dst,
			})
		}
	}

	// Create an independent component set for each end node of the graph.
	// Include the node + all their parent nodes. Only recompute the sets in
	// which at least one node has been modified.
	rawComponents := dd.FindComponents()

	components := [][]string{}
	for _, comp := range rawComponents {
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

	contracts := map[string]*solidity.Artifact{}

	// generate the outputs and compile
	for _, comp := range components {
		pragmas := []string{}
		for _, i := range comp {
			pragmas = append(pragmas, strings.Split(sources[i].Version[0], " ")...)
		}
		pragmas = unique(pragmas)

		// TODO: Parse this before
		versionConstraint, err := version.NewConstraint(strings.Join(pragmas, ", "))
		if err != nil {
			return err
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
		output, err := p.sol.Compile(input)
		if err != nil {
			return err
		}

		id := uuid.New()
		out := &SolcOutput{
			Id:     id.String(),
			Output: output,
		}
		p.state.Output[id.String()] = out

		for _, i := range comp {
			sources[i].BuildInfo = id.String()
		}
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

			// trim the lib directory from the path (if exists)
			path = strings.TrimPrefix(path, p.libDirectory)

			// remove the contracts path in the destination name
			sourcePath := filepath.Join(".greenhouse", path)
			if err := os.MkdirAll(sourcePath, 0755); err != nil {
				return err
			}

			// write the contract file
			raw, err := json.Marshal(contract)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(filepath.Join(sourcePath, name+".json"), raw, 0644); err != nil {
				return err
			}
		}
	}

	{
		// write metadata!
		// clean any extra output that is not being referenced anymore.
		deleteNames := []string{}
		for name := range p.state.Output {
			found := false
			for _, src := range p.state.Sources {
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
			delete(p.state.Output, name)
		}

		raw, err := json.Marshal(p.state)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(filepath.Join(".greenhouse", "metadata.json"), raw, 0644); err != nil {
			return err
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

func parsePragma(contract string) ([]string, error) {
	res := pragmaRegexp.FindStringSubmatch(contract)
	if len(res) == 0 {
		return nil, fmt.Errorf("pragma not found")
	}
	return res[1:], nil
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
