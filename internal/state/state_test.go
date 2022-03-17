package state

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestState_CreateSource(t *testing.T) {
	s, err := NewState()
	assert.NoError(t, err)

	src := &Source{
		Dir:      "./path",
		Filename: "contract.sol",
		ModTime:  time.Now(),
		Tainted:  true,
	}
	assert.NoError(t, s.UpsertSource(src))
	s.ListTaintedSources()
}
