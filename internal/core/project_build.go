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
	"github.com/umbracle/greenhouse/internal/state"
)

func (p *Project) findLocalDiff() error {
	files, err := Walk(p.config.Contracts)
	if err != nil {
		return err
	}

	sources, err := p.state.ListSources()
	if err != nil {
		panic(err)
	}
	diffFiles2, err := Diff1(sources, files)
	if err != nil {
		panic(err)
	}

	for _, diff := range diffFiles2 {
		if diff.Type == FileDiffAdd {

			src := diff.Source
			src.Tainted = true
			if err := p.state.UpsertSource(src); err != nil {
				panic(err)
			}
		}
		if diff.Type == FileDiffMod {
			// update the tainted
			//fmt.Println(diff.Source.Dir, diff.Source.Filename)
			if err := p.state.SetTaintedSource(diff.Source.Dir, diff.Source.Filename); err != nil {
				return err
			}
		}
	}
	return nil
}

// Compile compiles the application
func (p *Project) Compile() error {
	// solidity version we use to compile
	solidityVersion, err := version.NewVersion(p.config.Solidity)
	if err != nil {
		return err
	}

	updatedSources := []*state.Source{}

	sourcesList, err := p.state.ListSources()
	if err != nil {
		panic(err)
	}
	sources := map[string]*state.Source{} //p.state.Sources
	for _, s := range sourcesList {
		sources[s.Path()] = s

		if s.Tainted {
			updatedSources = append(updatedSources, s)
		}
	}

	remappings := map[string]string{}

	// detect dependencies only for new and modified files
	diffSources := []*state.Source{}
	for _, f := range updatedSources {
		content, err := ioutil.ReadFile(f.Path())
		if err != nil {
			return err
		}
		file, err := os.Stat(f.Path())
		if err != nil {
			fmt.Println(err)
		}

		imports := parseDependencies(string(content))
		parentPath := filepath.Dir(f.Path())

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
		/*
			src := &Source{
				Imports: cleanImports,
				Version: pragma,
				ModTime: f.ModTime,
				Path:    f.Path(),
				Hash:    hash(string(content)),
			}
		*/

		//fmt.Println("--mod time --")
		//fmt.Println(f.ModTime)

		src2 := f.Copy()
		src2.Tainted = false
		src2.Imports = cleanImports
		src2.Version = pragma
		src2.ModTime = file.ModTime()

		if err := p.state.UpsertSource(src2); err != nil {
			panic(err)
		}
		sources[f.Path()] = src2
		diffSources = append(diffSources, src2)
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
		//out := &SolcOutput{
		//	Id:     id.String(),
		//	Output: output,
		//}
		//p.state.Output[id.String()] = out

		for _, i := range comp {
			sources[i].BuildInfo = id.String()
		}
		for name, c := range output.Contracts {

			//fmt.Println("-- name --")
			//fmt.Println(name)

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
				return err
			}
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

	raw := writeMetadata(p.state)
	if err := ioutil.WriteFile(filepath.Join(".greenhouse", "metadata.json"), raw, 0644); err != nil {
		return err
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
