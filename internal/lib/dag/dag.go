package dag

import "sync"

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

	if s, ok := d.inbound[e.Src]; ok && s.(set).include(e.Dst) {
		return
	}

	s, ok := d.inbound[e.Src]
	if !ok {
		s = set{}
		d.inbound[e.Src] = s
	}
	s.(set).add(e.Dst)

	s, ok = d.outbound[e.Dst]
	if !ok {
		s = set{}
		d.outbound[e.Dst] = s
	}
	s.(set).add(e.Src)
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
