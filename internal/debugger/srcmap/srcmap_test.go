package srcmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSrcMap(t *testing.T) {
	cases := []struct {
		entry  string
		srcmap SrcMap
	}{
		{
			"1:2:1;:9;2:1:2;;",
			SrcMap{
				{0, 1, 2, 1, ""},
				{1, 1, 9, 1, ""},
				{2, 2, 1, 2, ""},
				{3, 2, 1, 2, ""},
				{4, 2, 1, 2, ""},
			},
		},
	}

	for _, c := range cases {
		srcmap, err := ParseSrcMap(c.entry)
		assert.NoError(t, err)
		assert.Equal(t, c.srcmap, srcmap)
	}
}
