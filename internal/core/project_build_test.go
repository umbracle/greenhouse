package core

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanImports(t *testing.T) {
	assert.Equal(t, "contract/b.sol", filepath.Join("contract/folder", "../b.sol"))
}

func TestParseImports(t *testing.T) {
	i := parseDependencies(`
		import "./a.sol";
		import './b.sol';
	`)
	assert.Equal(t, i[0], "./a.sol")
	assert.Equal(t, i[1], "./b.sol")
}
