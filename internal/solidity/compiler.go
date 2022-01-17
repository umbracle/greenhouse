package solidity

import "encoding/json"

type Compiler interface {
	Compile(i *Input) (*Output, error)
}

type Optimizer struct {
	Enabled bool
	Runs    int
}

type Input struct {
	Version    string
	Files      []string
	Optimizer  *Optimizer
	Remappings map[string]string
}

type Artifact struct {
	Abi           json.RawMessage `json:"abi"`
	Bin           string          `json:"bin"`
	BinRuntime    string          `json:"bin-runtime"`
	SrcMap        string          `json:"srcmap"`
	SrcMapRuntime string          `json:"srcmap-runtime"`
}

type Output struct {
	Contracts map[string]*Artifact
	Sources   map[string]*Source
	Version   string
}

type Source struct {
	AST *ASTNode
}
