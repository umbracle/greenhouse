package core

import (
	"fmt"
	"strings"

	"github.com/umbracle/greenhouse/internal/solidity"
)

func (p *Project) Graph() ([]byte, error) {
	if err := p.Compile(); err != nil {
		return nil, err
	}

	d := &dotMarshaler{
		p: p.metadata2,
	}
	res, err := d.Dot()
	if err != nil {
		return nil, err
	}
	return res, nil
}

type dotMarshaler struct {
	p *Metadata
}

func (d *dotMarshaler) Dot() ([]byte, error) {
	srcs := map[string][]string{}
	deps := map[string]string{}

	// build a map of contracts
	for _, i := range d.p.Output {
		for sourceName, j := range i.Output.Sources {

			contracts := []string{}

			fmt.Println("-- ast ")
			fmt.Println(j.AST)

			j.AST.Visit(solidity.ASTContractDefinition, func(n *solidity.ASTNode) {
				contractName := n.GetAttribute("name").(string)
				contracts = append(contracts, contractName)

				n.Visit(solidity.ASTInheritanceSpecifier, func(n *solidity.ASTNode) {
					fmt.Println("-- inheritance --")
					fmt.Println(n)

					for _, child := range n.Children {
						deps[contractName] = child.GetAttribute("name").(string)
					}
				})
			})
			srcs[sourceName] = contracts
		}
	}

	fmt.Println(srcs)
	fmt.Println(deps)

	str := "digraph G {\n"

	// create subgraphs for the sources
	count := 0
	for sourceName, contracts := range srcs {
		str += fmt.Sprintf("subgraph cluster_%d {\n", count)
		str += "node [style=filled];\n"
		str += "label = \"" + sourceName + "\";\n"
		str += fmt.Sprintf("\"%s\";\n", strings.Join(contracts, "\", \""))
		str += "}\n"
		count++
	}

	// create dependencies
	for from, to := range deps {
		str += fmt.Sprintf("\"%s\" -> \"%s\"\n", from, to)
	}

	str += "}"

	fmt.Println(str)
	return []byte(str), nil
}
