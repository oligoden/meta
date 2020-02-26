package refmap

import (
	"context"

	graph "github.com/oligoden/math-graph"
)

const (
	DataStable uint8 = iota
	DataChecked
	DataUpdated
	DataAdded
	DataRemove
)

type Store struct {
	reads  chan *readOp
	adds   chan *addOp
	Maps   chan *mapOp
	writes chan *writeOp
	Sets   chan *SetOp
	// sync    chan *SyncOp
	Changed chan *changedOp
	// Removed chan *RemovedOp
	// Exec    chan *ExecOp
	core  *core
	refs  map[string]Actioner
	maps  map[string]map[string]uint
	graph *graph.Graph
}

func Start(location string) *Store {
	s := &Store{}
	s.reads = make(chan *readOp)
	s.adds = make(chan *addOp)
	s.Maps = make(chan *mapOp)
	s.writes = make(chan *writeOp)
	s.Sets = make(chan *SetOp)
	// 	// s.sync = make(chan *SyncOp)
	s.Changed = make(chan *changedOp)
	// 	// s.Removed = make(chan *RemovedOp)
	// 	// s.Exec = make(chan *ExecOp)

	s.core = newCore(location)
	s.refs = make(map[string]Actioner)
	s.maps = make(map[string]map[string]uint)
	s.graph = graph.New()

	go func() {
		for {
			select {
			case a := <-s.reads:
				a.Rsp <- s.core.refs[a.Src]
			case a := <-s.writes:
				a.handle(s.core.root, s.core.refs)
			case a := <-s.adds:
				a.handle(s.refs, s.graph)
			case a := <-s.Maps:
				s.graph.Link(a.start, a.end)
				a.rsp <- nil
			case a := <-s.Sets:
				a.handle(s.core.refs, s.refs, s.graph)
			// case sync := <-s.sync:
			// 	sync.handle(s.core.refs)
			case changed := <-s.Changed:
				changed.handle(s.core.refs)
				// case removed := <-s.Removed:
				// 	removed.handle(s.core.refs)
				// case exec := <-s.Exec:
				// 	exec.handle(s.core.execs)
			}
		}
	}()

	return s
}

type core struct {
	root string
	refs map[string]*DestRef
	// execs map[string]action
}

func newCore(rl string) *core {
	c := &core{}
	c.root = rl
	c.refs = make(map[string]*DestRef)
	// c.execs = make(map[string]action)
	return c
}

// Actioner performs actions on the data provided.
type Actioner interface {
	Perform(context.Context) error
	Change(...uint8) uint8
	Hash() string
	// Remove(Config)
}

type Mutator interface {
	Write(string, string, Actioner)
}

type DestRef struct {
	Files  map[string]Actioner
	Change uint8
}

func NewDestination() *DestRef {
	r := &DestRef{}
	r.Files = make(map[string]Actioner)
	r.Change = DataAdded
	return r
}

var StatusText = map[uint8]string{
	DataStable:  "DataStable",
	DataChecked: "DataChecked",
	DataUpdated: "DataUpdated",
	DataAdded:   "DataAdded",
	DataRemove:  "DataRemove",
}
