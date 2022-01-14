package core

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	fmt.Println(version.NewConstraint(">=0.4.22, <0.6.0"))
	fmt.Println(version.NewConstraint(">=0.8.0"))
}

func TestCleanImports(t *testing.T) {

	fmt.Println(filepath.Join("contract/folder", "../b.sol"))
}

func TestParseImports(t *testing.T) {
	i := parseDependencies(`
		import "./a.sol";
		import './b.sol';
	`)
	assert.Equal(t, i[0], "./a.sol")
	assert.Equal(t, i[1], "./b.sol")
}
