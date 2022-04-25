package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/umbracle/greenhouse/internal/dag"
	"github.com/umbracle/greenhouse/internal/solidity"
	"github.com/umbracle/greenhouse/internal/state"
)

func (p *Project) findLocalDiff() error {
	files, err := Walk(p.config.Contracts)
	if err != nil {
		return err
	}

	sources, err := p.state.ListSources()
	if err != nil {
		return err
	}
	diffFiles2, err := Diff(sources, files)
	if err != nil {
		return err
	}

	for _, diff := range diffFiles2 {
		if diff.Type == FileDiffAdd {

			src := diff.Source
			src.Tainted = true

			if err := p.state.UpsertSource(src); err != nil {
				return err
			}
		}
		if diff.Type == FileDiffMod {
			// update the tainted
			if err := p.state.UpsertSource(diff.Source); err != nil {
				return err
			}
		}
	}
	return nil
}

// Compile compiles the application
func (p *Project) Compile() error {
	resp, err := p.compileImpl()
	if err != nil {
		return err
	}

	// write artifacts!
	for name, contract := range resp.Contracts {
		// name has the format <path>:<contract>
		// remove the contract name
		spl := strings.Split(name, ":")
		path, name := spl[0], spl[1]

		// trim the lib directory from the path (if exists)
		path = strings.TrimPrefix(path, p.libDirectory)

		// remove the contracts path in the destination name
		sourcePath := p.getFile(".greenhouse", path)

		fmt.Println("--- compile path ---", path, sourcePath)
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

	// write metadata
	metadataRaw, err := getMetadataRaw(p.state)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(p.getFile(".greenhouse", "metadata.json"), metadataRaw, 0644); err != nil {
		return err
	}
	return nil

}

type CompileResult struct {
	Contracts map[string]*state.Contract
}

func (p *Project) compileImpl() (*CompileResult, error) {
	// solidity version we use to compile
	solidityVersion, err := version.NewVersion(p.config.Solidity)
	if err != nil {
		return nil, err
	}

	updatedSources := []*state.Source{}

	sourcesList, err := p.state.ListSources()
	if err != nil {
		return nil, err
	}
	sources := map[string]*state.Source{}
	for _, s := range sourcesList {
		sources[s.Path()] = s

		if s.Tainted {
			updatedSources = append(updatedSources, s)
		}
	}

	remappings := map[string]string{}

	// detect dependencies only for new and modified files
	diffSources := updatedSources

	// find the necessary remappings
	for _, f := range updatedSources {
		for _, im := range f.GetRemappings() {
			// check if we can resolve it using the remappings
			if fullPath, exists := p.remappings[im]; exists {
				remappings[im] = fullPath
			} else {
				return nil, fmt.Errorf("absolute paths cannot be resolved yet")
			}
		}
	}

	// build dag map (move this to own repo)
	dd := &dag.Dag{}
	for _, f := range sources {
		dd.AddVertex(f)
	}
	// add edges
	for _, src := range sources {
		for _, dst := range src.GetLocalImports() {
			dst, ok := sources[dst]
			if !ok {
				panic(fmt.Errorf("BUG: elem in DAG not found: %s", dst.Path()))
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
				subComp = append(subComp, i.(*state.Source).Path())
			}
			components = append(components, subComp)
		}
	}

	contracts := map[string]*state.Contract{}

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
			return nil, err
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
			return nil, err
		}

		for _, i := range comp {
			src := sources[i].Copy()
			src.Tainted = false

			// update hte sources
			if err := p.state.UpsertSource(src); err != nil {
				return nil, err
			}
		}
		for name, c := range output.Contracts {
			parts := strings.Split(name, ":")

			dir, filename := filepath.Dir(parts[0]), filepath.Base(parts[0])
			contractName := parts[1]

			ctnr := &state.Contract{
				Name:       contractName,
				Dir:        dir,
				Filename:   filename,
				Abi:        string(c.Abi),
				Bin:        c.Bin,
				BinRuntime: c.BinRuntime,
			}
			if err := p.state.UpsertContract(ctnr); err != nil {
				return nil, err
			}
			contracts[name] = ctnr
		}
	}

	resp := &CompileResult{
		Contracts: contracts,
	}
	return resp, nil
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
