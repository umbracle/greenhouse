package solidity

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownloader(t *testing.T) {
	tmpDir, err := ioutil.TempDir("/tmp", "sol-down")
	assert.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	assert.NoError(t, downloadSolidity("0.8.0", tmpDir))
}
