package refmap

import (
	"context"

	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type Store struct {
	Adds    chan *addOp
	Maps    chan *mapOp
	Sets    chan *SetOp
	Changed chan *changedOp
	// Removed chan *RemovedOp
	refs  map[string]Actioner
	graph *graph.Graph
}

func Start() *Store {
	s := &Store{}
	s.Adds = make(chan *addOp)
	s.Maps = make(chan *mapOp)
	s.Sets = make(chan *SetOp)
	s.Changed = make(chan *changedOp)
	// 	// s.Removed = make(chan *RemovedOp)

	s.refs = make(map[string]Actioner)
	s.graph = graph.New()

	go func() {
		for {
			select {
			case a := <-s.Adds:
				a.handle(s.refs, s.graph)
			case a := <-s.Maps:
				a.rsp <- s.graph.Link(a.start, a.end)
			case a := <-s.Sets:
				a.handle(s.refs, s.graph)
			case changed := <-s.Changed:
				changed.handle(s.refs, s.graph)
				// case removed := <-s.Removed:
				// 	removed.handle(s.core.refs)
			}
		}
	}()

	return s
}

// Actioner performs actions on the data provided.
type Actioner interface {
	Perform(context.Context) error
	State(...uint8) uint8
	Hash() string
	Identifier() string
	// Remove(Config)
}

type Mutator interface {
	AddRef(string, Actioner)
	MapRef(string, string, ...uint)
}

var StatusText = map[uint8]string{
	state.Stable:  "Stable",
	state.Checked: "Checked",
	state.Updated: "Updated",
	state.Added:   "Added",
	state.Remove:  "Remove",
}
