package core

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/evmc/v10/bindings/go/evmc"
	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/wallet"
	state "github.com/umbracle/greenhouse/internal/runtime"
	"github.com/umbracle/greenhouse/internal/solidity"
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
	Output   *state.Output
}

func (p *Project) Test() ([]*TestOutput, error) {
	if err := p.Compile(); err != nil {
		return nil, err
	}

	targets := []*testTarget{}
	visited := map[string]struct{}{}

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

	txn := state.NewTransition(evmc.Istanbul, state.TxContext{}, &state.EmptyState{})

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
			})
		}
	}

	return result, nil
}
