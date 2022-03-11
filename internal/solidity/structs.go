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

type ASTType string

var (
	ASTContractDefinition   ASTType = "ContractDefinition"
	ASTInheritanceSpecifier ASTType = "InheritanceSpecifier"
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
