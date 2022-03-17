package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState_ValidateSchema(t *testing.T) {
	assert.NoError(t, dbSchema.Validate())
}
