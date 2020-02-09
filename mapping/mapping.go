package mapping

import (
	"context"
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
	writes chan *writeOp
	Sets   chan *SetOp
	// sync    chan *SyncOp
	Changed chan *changedOp
	// Removed chan *RemovedOp
	// Exec    chan *ExecOp
	core *core
}

func Start(location string) *Store {
	m := &Store{}
	m.reads = make(chan *readOp)
	m.writes = make(chan *writeOp)
	m.Sets = make(chan *SetOp)
	// 	// m.sync = make(chan *SyncOp)
	m.Changed = make(chan *changedOp)
	// 	// m.Removed = make(chan *RemovedOp)
	// 	// m.Exec = make(chan *ExecOp)

	m.core = newCore(location)

	go func() {
		for {
			select {
			case read := <-m.reads:
				read.Rsp <- m.core.refs[read.Src]
			case write := <-m.writes:
				write.handle(m.core.root, m.core.refs)
			case setter := <-m.Sets:
				setter.handle(&m.core.root, m.core.refs)
			// case sync := <-m.sync:
			// 	sync.handle(m.core.refs)
			case changed := <-m.Changed:
				changed.handle(m.core.refs)
				// case removed := <-m.Removed:
				// 	removed.handle(m.core.refs)
				// case exec := <-m.Exec:
				// 	exec.handle(m.core.execs)
			}
		}
	}()

	return m
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
