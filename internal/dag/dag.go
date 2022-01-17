package dag

import (
	"sync"
)

// Dag is a directly acyclic graph
type Dag struct {
	once   sync.Once
	vertex set

	inbound  set
	outbound set
}

// Hashable is the interface implemented by vertex objects
// that have a hash representation
type Hashable interface {
	Hash() interface{}
}

// Vertex is a vertex in the graph
type Vertex interface{}

// Edge is an edge between two vertex of the graph
type Edge struct {
	Src Vertex
	Dst Vertex
}

func (d *Dag) init() {
	d.once.Do(func() {
		d.vertex = set{}
		d.inbound = set{}
		d.outbound = set{}
	})
}

func (d *Dag) GetInbound(v Vertex) (res []Vertex) {
	vals, ok := d.inbound[v]
	if !ok {
		return
	}
	for k := range vals.(set) {
		res = append(res, k)
	}
	return
}

func (d *Dag) GetOutbound(v Vertex) (res []Vertex) {
	vals, ok := d.outbound[v]
	if !ok {
		return
	}
	for k := range vals.(set) {
		res = append(res, k)
	}
	return
}

// AddVertex adds a new vertex on the DAG
func (d *Dag) AddVertex(v Vertex) {
	d.init()
	d.vertex.add(v)
}

// AddEdge adds a new edge on the DAG
func (d *Dag) AddEdge(e Edge) {
	d.init()

	if s, ok := d.inbound[e.Dst]; ok && s.(set).include(e.Src) {
		return
	}

	s, ok := d.inbound[e.Dst]
	if !ok {
		s = set{}
		d.inbound[e.Dst] = s
	}
	s.(set).add(e.Src)

	s, ok = d.outbound[e.Src]
	if !ok {
		s = set{}
		d.outbound[e.Src] = s
	}
	s.(set).add(e.Dst)
}

func (d *Dag) FindComponents() [][]Vertex {

	// find components without any inbound
	leafVertex := []Vertex{}
	for v := range d.vertex {
		if _, ok := d.inbound[v]; !ok {
			leafVertex = append(leafVertex, v)
		}
	}

	result := [][]Vertex{}

	// follow each leaf vertex upwards to find the component
	for _, leaf := range leafVertex {
		component := []Vertex{}

		queue := []Vertex{leaf}
		for len(queue) != 0 {
			var item Vertex
			item, queue = queue[0], queue[1:]

			component = append(component, item)
			if outbound, ok := d.outbound[item]; ok {
				for v := range outbound.(set) {
					queue = append(queue, v)
				}
			}
		}
		result = append(result, component)
	}
	return result
}

type set map[interface{}]interface{}

func (s set) add(v Vertex) {
	k := v
	if h, ok := v.(Hashable); ok {
		k = h.Hash()
	}
	if _, ok := s[k]; !ok {
		s[k] = struct{}{}
	}
}

func (s set) include(v Vertex) bool {
	_, ok := s[v]
	return ok
}
