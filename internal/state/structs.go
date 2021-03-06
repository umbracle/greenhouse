package state

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/umbracle/ethgo/abi"
	"github.com/umbracle/greenhouse/internal/solidity"
)

type Source struct {
	// Dir is the directory of the file
	Dir string

	// Filename is the name of the file
	Filename string

	// ModTime is the modified time of the source
	ModTime time.Time

	// Tainted signals whether the code has been modified
	Tainted bool

	// Versions are the required version for this source
	Version []string

	// Imports is the list of imports defined in this source
	Imports []string

	// AST is **compiled** ast tree of the solidity code
	AST *solidity.ASTNode
}

func (s *Source) GetLocalImports() (local []string) {
	for _, i := range s.Imports {
		if strings.HasPrefix(i, "/") {
			local = append(local, i)
		}
	}
	return
}

func (s *Source) GetRemappings() (remappings []string) {
	for _, i := range s.Imports {
		if !strings.HasPrefix(i, "/") {
			remappings = append(remappings, i)
		}
	}
	return
}

func (s *Source) Path() string {
	return filepath.Join(s.Dir, s.Filename)
}

func (s *Source) Copy() *Source {
	ss := new(Source)
	*ss = *s
	return ss
}

type Contract struct {
	Dir      string // bundle this
	Filename string

	// Name is the name of the contract
	Name string

	// Abi is the abi encoding of the contract
	Abi string `json:"abi"`

	// Bin is the bin bytecode to deploy the contract
	Bin string `json:"bin"`

	// BinRuntime is the deployed bytecode of the contract
	BinRuntime string `json:"bin-runtime"`

	// SrcMap is the source map object for the deployment transaction
	SrcMap string `json:"srcmap"`

	// SrcMapRuntime is the source map object for the deployed contract
	SrcMapRuntime string `json:"srcmap-runtime"`
}

func (c *Contract) ABI() *abi.ABI {
	parsedAbi, err := abi.NewABI(string(c.Abi))
	if err != nil {
		panic(err)
	}
	return parsedAbi
}
