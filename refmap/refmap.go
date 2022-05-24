package refmap

import (
	"context"
	"fmt"
	"os"

	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type ContextKey string

type Store struct {
	Adds       chan *addOp
	Maps       chan *mapOp
	Sets       chan *SetOp
	Nodes      chan *nodesOp
	OutputChan chan struct{}
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
	s.OutputChan = make(chan struct{})
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
			case <-s.OutputChan:
				f, err := os.Create("output.gv")
				if err != nil {
					fmt.Println("error outputting graph ->", err)
					break
				}
				buf := s.graph.Output(
					"^prj", `[style=filled, fillcolor="slateblue1"]`,
					"^dir", `[style=filled, fillcolor="lightblue" shape="folder"]`,
					"^file", `[style=filled, fillcolor="lightgreen" shape="note"]`,
					"^exec", `[style=filled, fillcolor="lightcoral" shape="octagon"]`,
				)
				buf.WriteTo(f)
			}
		}
	}()

	return s
}

// Actioner performs actions on the data provided.
type Actioner interface {
	Perform(Grapher, context.Context) error
	State() uint8
	FlagState()
	ClearState()
	Hash() string
	Identifier() string
	Output() string
	// Remove(Config)
}

type Grapher interface {
	ParentFiles(string) []string
}

type Mutator interface {
	AddRef(context.Context, string, Actioner)
	MapRef(context.Context, string, string, ...uint) error
}

var StatusText = map[uint8]string{
	state.Stable:  "Stable",
	state.Checked: "Checked",
	state.Updated: "Updated",
	state.Added:   "Added",
	state.Remove:  "Remove",
}
