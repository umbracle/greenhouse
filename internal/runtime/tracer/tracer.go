package tracer

import (
	"math/big"

	"github.com/ethereum/evmc/v10/bindings/go/evmc"
)

type Tracer interface {
	CaptureState(pc int, op byte, gas, cost uint64, rData []byte, depth int, err error)
	CaptureStart(typ evmc.CallKind, from evmc.Address, to evmc.Address, input []byte, gas uint64, value *big.Int)
	CaptureEnd(ret []byte, gas int64, err error)
}

var _ Tracer = &StructTracer{}

// StructTracer is a Tracer implementation
type StructTracer struct {
	exec   *RunExec
	output *StructTracerOutput
}

type StructTracerOutput struct {
	Logs  []*StructLog
	Execs []*RunExec
}

func (s *StructTracer) GetOutput() *StructTracerOutput {
	return s.output
}

func (s *StructTracer) Reset() {
	s.output = &StructTracerOutput{
		Logs:  []*StructLog{},
		Execs: []*RunExec{},
	}
}

type RunExec struct {
	// input
	Typ   evmc.CallKind
	From  evmc.Address
	To    evmc.Address
	Input []byte
	Gas   uint64
	Value *big.Int

	// output
	Ret     []byte
	GasLeft int64
	Err     error
}

type StructLog struct {
	Pc      int
	Op      byte
	Gas     uint64
	Cost    uint64
	RetData []byte
	Depth   int
	Err     error
}

func (n *StructTracer) CaptureState(pc int, op byte, gas, cost uint64, rData []byte, depth int, err error) {
	log := &StructLog{
		Pc:      pc,
		Op:      op,
		Gas:     gas,
		Cost:    cost,
		RetData: append([]byte{}, rData...),
		Depth:   depth,
		Err:     err,
	}
	n.output.Logs = append(n.output.Logs, log)
}

func (n *StructTracer) CaptureStart(typ evmc.CallKind, from evmc.Address, to evmc.Address, input []byte, gas uint64, value *big.Int) {
	n.exec = &RunExec{
		Typ:   typ,
		From:  from,
		To:    to,
		Input: append([]byte{}, input...),
		Gas:   gas,
		Value: new(big.Int).Set(value),
	}
}

func (n *StructTracer) CaptureEnd(ret []byte, gas int64, err error) {
	n.exec.Ret = append([]byte{}, ret...)
	n.exec.GasLeft = gas
	n.exec.Err = err

	n.output.Execs = append(n.output.Execs, n.exec)
}
