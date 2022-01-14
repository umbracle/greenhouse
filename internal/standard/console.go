package standard

import (
	"encoding/hex"
	"regexp"
	"strings"

	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
)

var LogCases = map[string]*abi.Type{}

func init() {
	//fmt.Println("MAIN")

	/*
		content, err := ioutil.ReadFile("console.sol")
		if err != nil {
			// go generate ./internal/sol-test-lib
			panic(err)
		}
	*/

	rxp := regexp.MustCompile("abi.encodeWithSignature\\(\"log(.*)\"")
	matches := rxp.FindAllStringSubmatch(string(console), -1)

	for _, match := range matches {
		typStr := match[1]

		// TODO, fix web3 to work without the 256
		typStr = strings.Replace(typStr, "int,", "int256,", -1)
		typStr = strings.Replace(typStr, "int)", "int256)", -1)

		typ, err := abi.NewType("tuple" + typStr)
		if err != nil {
			panic(err)
		}

		// use the one without 256 for the sig
		xx := web3.Keccak256([]byte("log" + match[1]))
		xx = xx[:4]

		LogCases[hex.EncodeToString(xx)] = typ
	}
}
