package format

import (
	"fmt"

	solcparser "github.com/umbracle/solidity-parser-go"
)

type Opts struct {
}

func Format(code []byte) {
	p := solcparser.Parse(string(code))
	fmt.Println(p)

	t := solcparser.NewTreeSitter(string(code))
	fmt.Println(t)
}
