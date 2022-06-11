package refmap

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"

	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type ContextKey string

// Actioner performs actions on the data provided.
type Actioner interface {
	Perform(Grapher, context.Context) error
	State() uint8
	FlagState()
	ClearState()
	RemoveState()
	Hash() string
	Identifier() string
	Output() string
}

type Grapher interface {
	ParentFiles(string) []string
	Nodes(...string) []Actioner
}

type Mutator interface {
	AddRef(context.Context, string, Actioner)
	RenameRef(context.Context, string, string)
	MapRef(context.Context, string, string, ...uint) error
	Grapher
}

type Store struct {
	Adds       chan *addOp
	Rnms       chan *rnmOp
	Maps       chan *mapOp
	Sets       chan *SetOp
	Read       chan *readOp
	OutputChan chan struct{}
	// Removed chan *RemovedOp
	refs  map[string]Actioner
	graph *graph.Graph
}

func Start() *Store {
	s := &Store{}
	s.Adds = make(chan *addOp)
	s.Rnms = make(chan *rnmOp)
	s.Maps = make(chan *mapOp)
	s.Sets = make(chan *SetOp)
	s.Read = make(chan *readOp)
	s.OutputChan = make(chan struct{})

	s.refs = make(map[string]Actioner)
	s.graph = graph.New()

	go func() {
		for {
			select {
			case a := <-s.Adds:
				a.handle(s.refs, s.graph)
			case a := <-s.Rnms:
				a.handle(s.refs, s.graph)
			case a := <-s.Maps:
				// fmt.Println("linking", a.start, a.end)
				a.rsp <- s.graph.Link(a.start, a.end)
			case a := <-s.Sets:
				a.handle(s.refs, s.graph)
			case nodes := <-s.Read:
				if nodes.selection == "parents" {
					nodes.parents(nodes.node, s.refs, s.graph)
					break
				}
				nodes.topological(s.refs, s.graph)
			case <-s.OutputChan:
				f, err := os.Create("output.gv")
				if err != nil {
					fmt.Println("error outputting graph ->", err)
					break
				}

				buf := bytes.NewBufferString("digraph {\n")

				as := [][2]string{
					{"^prj", `[style=filled, fillcolor="slateblue1"]`},
					{"^dir", `[style=filled, fillcolor="lightblue" shape="folder"]`},
					{"^file", `[style=filled, fillcolor="lightgreen" shape="note"]`},
					{"^exec", `[style=filled, fillcolor="lightcoral" shape="octagon"]`},
				}

				nds, lks := s.graph.Graph()
				for _, name := range nds {
					if n, fnd := s.refs[name]; fnd {
						if n.State() == state.Remove {
							continue
						}
					}

					fmt.Fprintf(buf, "\t\"%s\"", name)
					for _, node := range as {
						match, _ := regexp.MatchString(node[0], name)
						if match {
							fmt.Fprintf(buf, " %s", node[1])
							break
						}
					}
					fmt.Fprintln(buf, ";")
				}

				for _, link := range lks {
					if n, fnd := s.refs[link[0]]; fnd {
						if n.State() == state.Remove {
							continue
						}
					}

					if n, fnd := s.refs[link[1]]; fnd {
						if n.State() == state.Remove {
							continue
						}
					}

					fmt.Fprintf(buf, "\t\"%s\" -> \"%s\";\n", link[0], link[1])
				}

				buf.WriteString("}")
				buf.WriteTo(f)
			}
		}
	}()

	return s
}
