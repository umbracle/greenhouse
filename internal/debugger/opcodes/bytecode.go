package opcodes

type Bytecode []*ParsedOpcode

type ParsedOpcode struct {
	OpCode OpCode
	Index  int
	PC     int
}

func (p *Bytecode) Index(i int) *ParsedOpcode {
	for _, o := range *p {
		if o.Index == i {
			return o
		}
	}
	return nil
}

// NewBytecode parses a bytecode in hex format and returns a
// Bytecode object
func NewBytecode(bytecode []byte) Bytecode {
	res := Bytecode{}

	byteIndex := 0
	instructionIndex := 0

	for byteIndex < len(bytecode) {
		instruction := bytecode[byteIndex]
		opcode := OpCode(instruction)

		length := 1
		if opcode.IsPush() {
			length = int(instruction - 0x60 + 2)
		}

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
