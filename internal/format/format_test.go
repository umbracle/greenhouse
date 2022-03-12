package format

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	data, err := ioutil.ReadFile("./testdata/format-contract-in.sol")
	assert.NoError(t, err)

	Format(data)
}
