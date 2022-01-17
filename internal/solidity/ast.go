package solidity

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
