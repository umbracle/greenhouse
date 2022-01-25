package standard

import _ "embed"

//go:embed console.sol
var console string

var SystemContracts = map[string]string{
	"greenhouse/console.sol": console,
}
