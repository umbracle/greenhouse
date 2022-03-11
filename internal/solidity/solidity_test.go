package solidity

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloader_Concurrent(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/tmp", "solc-test")
	assert.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	s := NewSolidity(tmpDir)
	assert.NoError(t, s.download("0.8.0"))
}
