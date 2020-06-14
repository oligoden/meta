package refmap

import (
	"strings"

	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type nodesOp struct {
	filter    string
	selection string
	node      string
	nodes     chan string
	Refs      chan Actioner
}

func (o nodesOp) topological(refs map[string]Actioner, g *graph.Graph) {
	g.CompileRun(func(ref string) error {
		if !strings.HasPrefix(ref, o.filter) {
			return nil
		}
		if refs[ref].State() == state.Updated || refs[ref].State() == state.Added {
			o.Refs <- refs[ref]
		}
		return nil
	})
	close(o.Refs)
}

func (o nodesOp) parents(node string, refs map[string]Actioner, g *graph.Graph) {
	g.ReverseRun(func(ref string) error {
		if !strings.HasPrefix(ref, o.filter) {
			return nil
		}
		o.nodes <- ref
		return nil
	}, node)
	close(o.nodes)
}

// ChangedRefs returns a slice of DestRefs that has changed.
func (r Store) ChangedRefs() []Actioner {
	refs := []Actioner{}
	changed := &nodesOp{
		selection: "changed",
		Refs:      make(chan Actioner),
	}
	r.Nodes <- changed

	for ref := range changed.Refs {
		refs = append(refs, ref)
	}
	return refs
}

// ChangedFiles returns a slice of files that has changed.
func (r Store) ChangedFiles() []Actioner {
	refs := []Actioner{}
	changed := &nodesOp{
		filter:    "file",
		selection: "changed",
		Refs:      make(chan Actioner),
	}
	r.Nodes <- changed

	for ref := range changed.Refs {
		refs = append(refs, ref)
	}
	return refs
}

// ChangedExecs returns a slice of execs that has changed.
func (r Store) ChangedExecs() []Actioner {
	refs := []Actioner{}
	changed := &nodesOp{
		filter:    "exec",
		selection: "changed",
		Refs:      make(chan Actioner),
	}
	r.Nodes <- changed

	for ref := range changed.Refs {
		refs = append(refs, ref)
	}
	return refs
}

// ParentFiles returns a slice of all the parent files.
func (r Store) ParentFiles(file string) []string {
	parents := &nodesOp{
		filter:    "file",
		selection: "parents",
		node:      file,
		nodes:     make(chan string),
	}
	r.Nodes <- parents

	nodes := []string{}
	for node := range parents.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}
