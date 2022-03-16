package core

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/evmc/v10/bindings/go/evmc"
	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/wallet"
	state "github.com/umbracle/greenhouse/internal/runtime"
	"github.com/umbracle/greenhouse/internal/solidity"
	"github.com/umbracle/greenhouse/internal/standard"
)

type testTarget struct {
	Source   string
	Name     string
	Addr     web3.Address
	Abi      *abi.ABI
	Artifact *solidity.Artifact
}

type TestOutput struct {
	Source   string
	Contract string
	Method   string
	Console  []*ConsoleOutput
	Output   *state.Output
}

type TestInput struct {
	Run string
}

func (p *Project) Test(input *TestInput) ([]*TestOutput, error) {
	if err := p.Compile(); err != nil {
		return nil, err
	}

	targets := []*testTarget{}
	visited := map[string]struct{}{}

	runExpr, err := regexp.Compile(input.Run)
	if err != nil {
		return nil, fmt.Errorf("failed to decode 'run' regexp expr: %v", err)
	}
	isValidFunc := func(name string) bool {
		return runExpr.Match([]byte(name))
	}

	for _, i := range p.state.Sources {
		out := p.state.Output[i.BuildInfo]
		for complexName, c := range out.Output.Contracts {

			// contract names might be repeated
			if _, ok := visited[complexName]; ok {
				continue
			}
			visited[complexName] = struct{}{}

			parts := strings.Split(complexName, ":")
			sourceName, contractName := parts[0], parts[1]
			if !strings.HasPrefix(contractName, "Test") {
				continue
			}

			contractABI, err := abi.NewABI(string(c.Abi))
			if err != nil {
				return nil, err
			}

			// check if there is any contract that matches the regexp
			validContract := false
			for method := range contractABI.Methods {
				if !strings.HasPrefix(method, "test") {
					continue
				}
				if isValidFunc(method) {
					validContract = true
					break
				}
			}
			if !validContract {
				continue
			}

			targets = append(targets, &testTarget{
				Source:   sourceName,
				Name:     contractName,
				Abi:      contractABI,
				Artifact: c,
			})
		}
	}

	// address to deploy the contracts
	key, _ := wallet.GenerateKey()
	sender := key.Address()

	console := &consoleCheatcode{}
	console.reset()

	opts := []state.ConfigOption{
		state.WithRevision(evmc.Istanbul),
		state.WithCheatcode(console),
	}
	txn := state.NewTransition(opts...)

	targetsByAddr := map[web3.Address]*testTarget{}
	for _, target := range targets {
		code, err := hex.DecodeString(target.Artifact.Bin)
		if err != nil {
			return nil, err
		}
		bin, err := hex.DecodeString(target.Artifact.BinRuntime)
		if err != nil {
			return nil, err
		}

		// deploy the contract
		msg := &state.Message{GasPrice: big.NewInt(1), Gas: 1000000000, From: evmc.Address(sender), To: nil, Input: code, Value: big.NewInt(0)}
		output := txn.Apply(msg)
		if !output.Success {
			return nil, fmt.Errorf("failed to deploy")
		}

		// check deployed code
		deployedCode := txn.GetCode(output.ContractAddress)
		if !bytes.Equal(deployedCode, bin) {
			return nil, fmt.Errorf("deployed code does not match")
		}

		target.Addr = web3.Address(output.ContractAddress)
		targetsByAddr[target.Addr] = target
	}

	result := []*TestOutput{}
	for _, target := range targets {
		for method, sig := range target.Abi.Methods {
			if !strings.HasPrefix(method, "test") {
				continue
			}
			if !isValidFunc(method) {
				continue
			}

			to := evmc.Address(target.Addr)
			msg := &state.Message{
				GasPrice: big.NewInt(1),
				Value:    big.NewInt(0),
				Gas:      1000000000,
				From:     evmc.Address(sender),
				To:       &to,
				Input:    sig.ID(),
			}
			output := txn.Apply(msg)

			result = append(result, &TestOutput{
				Source:   target.Source,
				Contract: target.Name,
				Method:   method,
				Output:   output,
				Console:  console.outputs,
			})
			console.reset()
		}
	}

	return result, nil
}

type ConsoleOutput struct {
	Err error
	Val []string
}

type consoleCheatcode struct {
	outputs []*ConsoleOutput
}

func (c *consoleCheatcode) CanRun(addr evmc.Address) bool {
	return hex.EncodeToString(addr[:]) == "000000000000000000636f6e736f6c652e6c6f67"
}

func (c *consoleCheatcode) reset() {
	c.outputs = []*ConsoleOutput{}
}

func (c *consoleCheatcode) addError(err error) {
	c.outputs = append(c.outputs, &ConsoleOutput{Err: err})
}

func (c *consoleCheatcode) Run(addr evmc.Address, input []byte) {
	sig := hex.EncodeToString(input[:4])
	logSig, ok := standard.LogCases[sig]
	if !ok {
		c.addError(fmt.Errorf("sig %s not found", sig))
		return
	}
	input = input[4:]
	raw, err := logSig.Decode(input)
	if err != nil {
		c.addError(fmt.Errorf("failed to decode: %v", err))
		return
	}
	val := []string{}
	for _, v := range raw.(map[string]interface{}) {
		val = append(val, fmt.Sprint(v))
	}
	c.outputs = append(c.outputs, &ConsoleOutput{
		Val: val,
	})
}
