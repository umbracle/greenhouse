package core

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
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
	soltestlib "github.com/umbracle/greenhouse/internal/sol-test-lib"
)

type testTarget struct {
	Artifact string
	File     string
	Contract *Contract
	Addr     web3.Address
	Abi      *abi.ABI
	SrcMap   *SrcMap
}

type TestInput struct {
	Prefix string
}

func (p *Project) Test(input *TestInput) error {
	if err := p.Compile(); err != nil {
		panic(err)
	}

	// Find test contracts
	targets := []*testTarget{}
	for file, i := range p.metadata {
		for _, j := range i.Artifacts {
			if strings.HasPrefix(j, "Test") {

				a, err := abi.NewABI(i.Contract[0].Artifact.Abi)
				if err != nil {
					panic(err)
				}

				targets = append(targets, &testTarget{
					Artifact: j,
					File:     file,
					Contract: i.Contract[0], // this only works because there is one, fix
					Abi:      a,
				})
			}
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

	tracer := &stateTracer{}
	transition := state.NewTransition(forks, config, snap, &state.Config{Tracer: tracer})

	// append custom runtimes
	consoleRuntime := &consoleRuntime{
		addr: web3.HexToAddress("0x000000000000000000636F6e736F6c652e6c6f67"),
	}
	transition.SetRuntime(consoleRuntime)

	//testRuntime := &testRuntime{}
	//transition.SetRuntime(testRuntime)

	targetsByAddr := map[web3.Address]*testTarget{}

	for _, target := range targets {
		code, err := hex.DecodeString(target.Contract.Artifact.Bin)
		if err != nil {
			panic(err)
		}

		bin, err := hex.DecodeString(target.Contract.Artifact.BinRuntime)
		if err != nil {
			panic(err)
		}

		res := transition.Create(sender, code, big.NewInt(0), 1000000000)
		if res.Err != nil {
			panic(res.Err)
		}
		if !bytes.Equal(res.ReturnValue, bin) {
			panic("bad")
		}
		target.Addr = web3.BytesToAddress(res.CreateAddress.Bytes())

		srcmap, err := ParseSrcMap(target.Contract.Artifact.SrcMapRuntime)
		if err != nil {
			panic(err)
		}
		target.SrcMap = srcmap
		targetsByAddr[target.Addr] = target
	}

	// execute the functions
	for _, target := range targets {
		for method, sig := range target.Abi.Methods {
			tracer.reset()

			if !strings.HasPrefix(method, "test") {
				continue
			}

			if input.Prefix != "" {
				if !strings.HasPrefix(method, input.Prefix) {
					continue
				}
			}

			fmt.Println("-- call --")
			fmt.Println(transition.Call(sender, types.Address(target.Addr), sig.ID(), big.NewInt(0), 10000000000000000))
			fmt.Println(transition.Txn().Logs())
			fmt.Println("-- trace --")

			data, err := json.Marshal(tracer.trace)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(data))
		}
	}

	return nil
}

type consoleRuntime struct {
	addr web3.Address
}

func (cc *consoleRuntime) Run(c *runtime.Contract, host runtime.Host, config *runtime.ForksInTime, tracer runtime.Tracer) *runtime.ExecutionResult {
	sig := hex.EncodeToString(c.Input[:4])
	logSig, ok := soltestlib.LogCases[sig]
	if !ok {
		panic("bad")
	}

	input := c.Input[4:]
	fmt.Println(logSig.Decode(input))

	return &runtime.ExecutionResult{}
}

func (cc *consoleRuntime) CanRun(c *runtime.Contract, host runtime.Host, config *runtime.ForksInTime) bool {
	return cc.addr == web3.Address(c.CodeAddress)
}

func (cc *consoleRuntime) Name() string {
	return "console"
}

type traceSimple struct {
	Pc    int  `json:"pc"`
	Depth int  `json:"depth"`
	Op    byte `json:"opcode"`
}

type traceStep struct {
	Call *rawTrace    `json:"call,omitempty"`
	Step *traceSimple `json:"step,omitempty"`
}

type rawTrace struct {
	parent *rawTrace

	Addr types.Address
	Sig  []byte

	Steps []*traceStep
}

type stateTracer struct {
	trace *rawTrace
}

func (s *stateTracer) reset() {
	s.trace = nil
}

func (s *stateTracer) TraceStart(sig []byte, addr types.Address) {
	trace := &rawTrace{
		parent: s.trace,
		Addr:   addr,
		Sig:    sig,
		Steps:  []*traceStep{},
	}
	if s.trace == nil {
		s.trace = trace
	} else {
		s.trace.Steps = append(s.trace.Steps, &traceStep{
			Call: trace,
		})
		s.trace = trace
	}
}

func (s *stateTracer) TraceEnd() {
	if s.trace.parent != nil {
		s.trace = s.trace.parent
	}
}

func (s *stateTracer) TraceState(pc, depth int, op byte) {
	s.trace.Steps = append(s.trace.Steps, &traceStep{
		Step: &traceSimple{
			Pc:    pc,
			Depth: depth,
			Op:    op,
		},
	})
}

func (s *stateTracer) TraceError(err error) {

}

/*
// test runtime

type testRuntime struct {
	addr web3.Address
}

func (tt *testRuntime) Run(c *runtime.Contract, host runtime.Host, config *runtime.ForksInTime) *runtime.ExecutionResult {
	return nil
}

func (tt *testRuntime) CanRun(c *runtime.Contract, host runtime.Host, config *runtime.ForksInTime) bool {
	fmt.Println("_ TEST __")
	fmt.Println(c.CodeAddress)

	return false
}

func (tt *testRuntime) Name() string {
	return "test"
}
*/
