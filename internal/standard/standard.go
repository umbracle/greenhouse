package standard

//go:embed console.sol
var console string

var SystemContracts = map[string]string{
	"greenhouse/console.sol": console,
}
