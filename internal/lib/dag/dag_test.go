package dag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDag_InOutbound(t *testing.T) {
	d := &Dag{}
	d.AddVertex(1)
	d.AddVertex(2)

	d.AddEdge(Edge{
		Src: 1,
		Dst: 2,
	})

	assert.Equal(t, d.GetOutbound(1)[0], 2)
	assert.Equal(t, d.GetInbound(2)[0], 1)
}

func TestDag_FindComponents(t *testing.T) {
	d := &Dag{}
	d.AddVertex(1)
	d.AddVertex(2)
	d.AddVertex(3)

	d.AddEdge(Edge{
		Src: 2,
		Dst: 1,
	})
	d.AddEdge(Edge{
		Src: 3,
		Dst: 1,
	})

	d.FindComponents()
}
