package opcodes

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytecode(t *testing.T) {
	cases := []struct {
		str      string
		bytecode Bytecode
	}{
		{
			"0102",
			Bytecode{
				{ADD, 0, 0},
				{MUL, 1, 1},
			},
		},
		{
			"60006000",
			Bytecode{
				{PUSH1, 0, 0},
				{PUSH1, 1, 2},
			},
		},
	}

	for _, c := range cases {
		buf, err := hex.DecodeString(c.str)
		assert.NoError(t, err)

		bytecode := NewBytecode(buf)
		assert.Equal(t, c.bytecode, bytecode)
	}
}
