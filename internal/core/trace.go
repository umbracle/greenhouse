package core

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/0xPolygon/eth-state-transition/runtime/evm"
)

type SrcMap []*SourceLocation

type SourceLocation struct {
	Index     int
	Offset    int
	Length    int
	FileIndex int
	Jump      string
}

func (s *SourceLocation) Src() string {
	return fmt.Sprintf("%d:%d:%d", s.Offset, s.Length, s.FileIndex)
}

func (s *SourceLocation) String() string {
	return fmt.Sprintf("Index: %d, Offset: %d, Length: %d, FileIndex: %d, Jump: %s", s.Index, s.Offset, s.Length, s.FileIndex, s.Jump)
}

func (s *SourceLocation) Copy() *SourceLocation {
	ss := new(SourceLocation)
	*ss = *s
	return ss
}

func ParseSrcMap(m string) (*SrcMap, error) {
	res := &SrcMap{}

	parseInt := func(s string) (int, error) {
		return strconv.Atoi(s)
	}

	var err error
	lastEntry := &SourceLocation{}

	for indx, entry := range strings.Split(m, ";") {
		current := strings.Split(entry, ":")

		var num int
		if len(entry) != 0 {
			num = len(current)
		}

		if num >= 1 && current[0] != "" {
			// offset
			if lastEntry.Offset, err = parseInt(current[0]); err != nil {
				return nil, err
			}
		}
		if num >= 2 && current[1] != "" {
			// length
			if lastEntry.Length, err = parseInt(current[1]); err != nil {
				return nil, err
			}
		}
		if num >= 3 && current[2] != "" {
			// file index
			if lastEntry.FileIndex, err = parseInt(current[2]); err != nil {
				return nil, err
			}
		}
		if num >= 4 && current[3] != "" {
			lastEntry.Jump = current[3]
		}

		entry := lastEntry.Copy()
		entry.Index = indx
		*res = append(*res, entry)
	}
	return res, nil
}

type ParsedBytecode []*ParsedOpcode

type ParsedOpcode struct {
	OpCode evm.OpCode
	PC     int
	Index  int
}

func (p *ParsedBytecode) Index(i int) *ParsedOpcode {
	for _, o := range *p {
		if o.Index == i {
			return o
		}
	}
	return nil
}

func IsPush(op evm.OpCode) bool {
	return strings.HasPrefix(op.String(), "PUSH")
}

func parseOpcodes(bytecodeHex string) ParsedBytecode {
	res := ParsedBytecode{}

	bytecode, err := hex.DecodeString(bytecodeHex)
	if err != nil {
		panic(err)
	}

	byteIndex := 0
	instructionIndex := 0

	for byteIndex < len(bytecode) {
		instruction := bytecode[byteIndex]
		opcode := evm.OpCode(instruction)

		length := 1
		if IsPush(opcode) {
			length = int(instruction - 0x60 + 2)
			// fmt.Println(length)
		}

		//if opcode.GoString() == "" {
		//	byteIndex += length
		//	continue
		//}

		//fmt.Printf("Opcode: %s %d %d\n", opcode.GoString(), byteIndex, instructionIndex)

		res = append(res, &ParsedOpcode{
			OpCode: opcode,
			PC:     byteIndex,
			Index:  instructionIndex,
		})
		byteIndex += length
		instructionIndex += 1
	}

	return res
}
