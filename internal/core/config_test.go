package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Merge(t *testing.T) {
	cfg := DefaultConfig()

	cfg2 := &Config{
		Solidity: "0.5.0",
	}
	assert.NoError(t, cfg.Merge(cfg2))
	assert.Equal(t, "0.5.0", cfg.Solidity)
}
