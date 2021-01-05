package refmap

import (
	"context"

	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type Store struct {
	Adds  chan *addOp
	Maps  chan *mapOp
	Sets  chan *SetOp
	Nodes chan *nodesOp
	// Removed chan *RemovedOp
	refs  map[string]Actioner
	graph *graph.Graph
}

func Start() *Store {
	s := &Store{}
	s.Adds = make(chan *addOp)
	s.Maps = make(chan *mapOp)
	s.Sets = make(chan *SetOp)
	s.Nodes = make(chan *nodesOp)
	// 	// s.Removed = make(chan *RemovedOp)

	s.refs = make(map[string]Actioner)
	s.graph = graph.New()

	go func() {
		for {
			select {
			case a := <-s.Adds:
				a.handle(s.refs, s.graph)
			case a := <-s.Maps:
				// fmt.Println("linking", a.start, a.end)
				a.rsp <- s.graph.Link(a.start, a.end)
			case a := <-s.Sets:
				a.handle(s.refs, s.graph)
			case nodes := <-s.Nodes:
				if nodes.selection == "parents" {
					nodes.parents(nodes.node, s.refs, s.graph)
					break
				}
				nodes.topological(s.refs, s.graph)
				// case removed := <-s.Removed:
				// 	removed.handle(s.core.refs)
			}
		}
	}()

	return s
}

// Actioner performs actions on the data provided.
type Actioner interface {
	Perform(Grapher, context.Context) error
	State(...uint8) uint8
	Hash() string
	Identifier() string
	Output() string
	// Remove(Config)
}

type Grapher interface {
	ParentFiles(string) []string
}

type Mutator interface {
	AddRef(string, Actioner)
	MapRef(string, string, ...uint) error
}

var StatusText = map[uint8]string{
	state.Stable:  "Stable",
	state.Checked: "Checked",
	state.Updated: "Updated",
	state.Added:   "Added",
	state.Remove:  "Remove",
}
