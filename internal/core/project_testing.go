package core

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	state "github.com/0xPolygon/eth-state-transition"
	itrie "github.com/0xPolygon/eth-state-transition/immutable-trie"
	"github.com/0xPolygon/eth-state-transition/runtime"
	"github.com/0xPolygon/eth-state-transition/types"
	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/wallet"
	"github.com/umbracle/greenhouse/internal/solidity"
	standard "github.com/umbracle/greenhouse/internal/standard"
)

type testTarget struct {
	Source   string
	Name     string
	Addr     web3.Address
	Abi      *abi.ABI
	Artifact *solidity.Artifact
}

type TestInput struct {
	Path    string
	Prefix  string
	Verbose bool
}

func (p *Project) Test(input *TestInput) error {
	if err := p.Compile(); err != nil {
		panic(err)
	}

	fmt.Println("-- input --")

	// Find test contracts
	targets := []*testTarget{}
	fmt.Println(p.metadata2)

	visited := map[string]struct{}{}
	for _, i := range p.metadata2.Sources {
		out := p.metadata2.Output[i.BuildInfo]
		for xx, c := range out.Output.Contracts {

			// contract names might be repeated
			if _, ok := visited[xx]; ok {
				continue
			}
			visited[xx] = struct{}{}

			contractName := strings.Split(xx, ":")[1]
			if !strings.HasPrefix(contractName, "Test") {
				continue
			}

			a, err := abi.NewABI(string(c.Abi))
			if err != nil {
				panic(err)
			}

			targets = append(targets, &testTarget{
				Name:     contractName,
				Abi:      a,
				Artifact: c,
			})
			fmt.Println(a, err)
		}

	}

	key, _ := wallet.GenerateKey()
	sender := types.BytesToAddress(key.Address().Bytes())

	// create the transition objects
	s := itrie.NewArchiveState(itrie.NewMemoryStorage())
	snap := s.NewSnapshot()

	forks := runtime.ForksInTime{
		Homestead:      true,
		Byzantium:      true,
		Constantinople: true,
		Petersburg:     true,
		Istanbul:       true,
		EIP150:         true,
		EIP158:         true,
		EIP155:         true,
	}
	config := runtime.TxContext{}

	transition := state.NewTransition(forks, config, snap)

	// append custom runtimes
	consoleRuntime := &consoleRuntime{
		addr: web3.HexToAddress("0x000000000000000000636F6e736F6c652e6c6f67"),
	}
	transition.SetRuntime(consoleRuntime)

	//testRuntime := &testRuntime{}
	//transition.SetRuntime(testRuntime)

	targetsByAddr := map[web3.Address]*testTarget{}

	for _, target := range targets {
		code, err := hex.DecodeString(target.Artifact.Bin)
		if err != nil {
			panic(err)
		}

		bin, err := hex.DecodeString(target.Artifact.BinRuntime)
		if err != nil {
			panic(err)
		}

		res := transition.Create(sender, code, big.NewInt(0), 1000000000)
		if res.Err != nil {
			panic(res.Err)
		}
		if !bytes.Equal(res.ReturnValue, bin) {
			// if the contract is created this should match
			panic("bad")
		}
		target.Addr = web3.BytesToAddress(res.CreateAddress.Bytes())
		targetsByAddr[target.Addr] = target
	}

	// fmt.Println(targets)

	// execute the functions
	for _, target := range targets {
		for method, sig := range target.Abi.Methods {
			if !strings.HasPrefix(method, "test") {
				continue
			}

			if input.Prefix != "" {
				if !strings.HasPrefix(method, input.Prefix) {
					continue
				}
			}

			fmt.Printf("=== RUN %s:%s\n", target.Name, method)
			//fmt.Println(target.Name + ":" + method)
			//fmt.Println("-- call --")
			//fmt.Println(method)
			transition.Call(sender, types.Address(target.Addr), sig.ID(), big.NewInt(0), 10000000000000000)
			//fmt.Println(transition.Txn().Logs())
			if input.Verbose {
				for _, val := range consoleRuntime.vals {
					fmt.Println(val)
				}
			}
			//fmt.Println("-- trace --")
		}
	}

	return nil
}

type consoleRuntime struct {
	addr web3.Address

	vals []string
}

func (cc *consoleRuntime) reset() {
	cc.vals = []string{}
}

func (cc *consoleRuntime) Run(c *runtime.Contract, host runtime.Host, config *runtime.ForksInTime) *runtime.ExecutionResult {
	sig := hex.EncodeToString(c.Input[:4])
	logSig, ok := standard.LogCases[sig]
	if !ok {
		panic("bad")
	}

	input := c.Input[4:]
	raw, err := logSig.Decode(input)
	if err != nil {
		panic("something bad we have to log")
	}

	val := []string{}
	for _, v := range raw.(map[string]interface{}) {
		val = append(val, fmt.Sprint(v))
	}
	cc.vals = append(cc.vals, strings.Join(val, " "))

	return &runtime.ExecutionResult{}
}

func (cc *consoleRuntime) CanRun(c *runtime.Contract, host runtime.Host, config *runtime.ForksInTime) bool {
	return cc.addr == web3.Address(c.CodeAddress)
}

func (cc *consoleRuntime) Name() string {
	return "console"
}
