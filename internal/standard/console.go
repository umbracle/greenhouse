package standard

import (
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
)

var LogCases = map[string]*abi.Type{}

func init() {
	rxp := regexp.MustCompile("abi.encodeWithSignature\\(\"log(.*)\"")
	matches := rxp.FindAllStringSubmatch(string(console), -1)

	for _, match := range matches {
		signature := match[1]

		// parse the type of the console call. Note that 'uint'
		// objects are defined without bytes (i.e. 256).
		typ, err := abi.NewType("tuple" + signature)
		if err != nil {
			panic(fmt.Errorf("BUG: Failed to parse %s", signature))
		}

		// signature of the call. Use the version without the bytes in 'uint'.
		sig := web3.Keccak256([]byte("log" + match[1]))[:4]
		LogCases[hex.EncodeToString(sig)] = typ
	}
}
