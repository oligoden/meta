package refmap

import (
	"context"

	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type Store struct {
	// reads chan *readOp
	Adds chan *addOp
	Maps chan *mapOp
	Sets chan *SetOp
	// sync    chan *SyncOp
	Changed chan *changedOp
	// Removed chan *RemovedOp
	// Exec    chan *ExecOp
	refs  map[string]Actioner
	graph *graph.Graph
}

func Start() *Store {
	s := &Store{}
	// s.reads = make(chan *readOp)
	s.Adds = make(chan *addOp)
	s.Maps = make(chan *mapOp)
	s.Sets = make(chan *SetOp)
	// 	// s.sync = make(chan *SyncOp)
	s.Changed = make(chan *changedOp)
	// 	// s.Removed = make(chan *RemovedOp)
	// 	// s.Exec = make(chan *ExecOp)

	s.refs = make(map[string]Actioner)
	s.graph = graph.New()

	go func() {
		for {
			select {
			// case a := <-s.reads:
			// 	a.Rsp <- s.core.refs[a.Src]
			case a := <-s.Adds:
				a.handle(s.refs, s.graph)
			case a := <-s.Maps:
				s.graph.Link(a.start, a.end)
				a.rsp <- nil
			case a := <-s.Sets:
				a.handle(s.refs, s.graph)
			// case sync := <-s.sync:
			// 	sync.handle(s.core.refs)
			case changed := <-s.Changed:
				changed.handle(s.refs, s.graph)
				// case removed := <-s.Removed:
				// 	removed.handle(s.core.refs)
				// case exec := <-s.Exec:
				// 	exec.handle(s.core.execs)
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
