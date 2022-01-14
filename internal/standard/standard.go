package standard

import _ "embed"

//go:embed assert.sol
var assert string

//go:embed console.sol
var console string

var SystemContracts = map[string]string{
	"greenhouse/assert.sol":  assert,
	"greenhouse/console.sol": console,
}
