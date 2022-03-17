package core

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strings"

	"github.com/ethereum/evmc/v10/bindings/go/evmc"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/wallet"
	state "github.com/umbracle/greenhouse/internal/runtime"
	"github.com/umbracle/greenhouse/internal/standard"
	state2 "github.com/umbracle/greenhouse/internal/state"
)

type testTarget struct {
	Source   string
	Name     string
	Addr     ethgo.Address
	Abi      *abi.ABI
	Contract *state2.Contract
}

type testTargets []*testTarget

func (t testTargets) Len() int {
	return len(t)
}

func (t testTargets) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t testTargets) Less(i, j int) bool {
	if t[i].Source != t[j].Source {
		return t[i].Source < t[j].Source
	}
	return t[i].Name < t[j].Name
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
	targets := testTargets{}

	runExpr, err := regexp.Compile(input.Run)
	if err != nil {
		return nil, fmt.Errorf("failed to decode 'run' regexp expr: %v", err)
	}
	isValidFunc := func(name string) bool {
		return runExpr.Match([]byte(name))
	}
	contracts, err := p.state.ListContracts()
	if err != nil {
		return nil, err
	}
	for _, contract := range contracts {
		if !strings.HasPrefix(contract.Name, "Test") {
			continue
		}

		contractABI, err := abi.NewABI(string(contract.Abi))
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
			Source:   contract.Dir + "/" + contract.Filename,
			Name:     contract.Name,
			Abi:      contractABI,
			Contract: contract,
		})
	}

	sort.Sort(targets)

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

	targetsByAddr := map[ethgo.Address]*testTarget{}
	for _, target := range targets {
		code, err := hex.DecodeString(target.Contract.Bin)
		if err != nil {
			return nil, err
		}
		bin, err := hex.DecodeString(target.Contract.BinRuntime)
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

		target.Addr = ethgo.Address(output.ContractAddress)
		targetsByAddr[target.Addr] = target
	}

	result := []*TestOutput{}
	for _, target := range targets {

		// sort the methods to get a deterministic output
		methodNames := sort.StringSlice{}
		for method, _ := range target.Abi.Methods {
			methodNames = append(methodNames, method)
		}
		sort.Sort(methodNames)

		for _, methodName := range methodNames {
			method := target.Abi.Methods[methodName]

			if !strings.HasPrefix(methodName, "test") {
				continue
			}
			if !isValidFunc(methodName) {
				continue
			}

			to := evmc.Address(target.Addr)
			msg := &state.Message{
				GasPrice: big.NewInt(1),
				Value:    big.NewInt(0),
				Gas:      1000000000,
				From:     evmc.Address(sender),
				To:       &to,
				Input:    method.ID(),
			}
			output := txn.Apply(msg)

			result = append(result, &TestOutput{
				Source:   target.Source,
				Contract: target.Name,
				Method:   methodName,
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
