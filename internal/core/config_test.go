package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Merge(t *testing.T) {
	cfg := DefaultConfig()

	cfg2 := &Config{
		Solidity: "0.5.0",
	}
	assert.NoError(t, cfg.Merge(cfg2))

	fmt.Println(cfg.Solidity)
}
