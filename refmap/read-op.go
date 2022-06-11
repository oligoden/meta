package refmap

import (
	"strings"

	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type readOp struct {
	filter    string
	selection string
	node      string
	nodes     chan string
	Refs      chan Actioner
}

func (o readOp) topological(refs map[string]Actioner, g *graph.Graph) {
	// fmt.Printf("\n\n%+v\n\n", g)
	// fmt.Printf("\n\n%+v\n\n", g.StartNodes())

	g.CompileRun(func(ref string) error {
		if o.filter != "" && !strings.HasPrefix(ref, o.filter) {
			return nil
		}
		if o.selection == "changed" &&
			(refs[ref].State() == state.Updated ||
				refs[ref].State() == state.Added) {
			o.Refs <- refs[ref]
			return nil
		}
		o.Refs <- refs[ref]
		return nil
	})
	close(o.Refs)
}

func (o readOp) parents(node string, refs map[string]Actioner, g *graph.Graph) {
	g.ReverseRun(func(ref string) error {
		if !strings.HasPrefix(ref, o.filter) {
			return nil
		}
		o.nodes <- ref
		return nil
	}, node)
	close(o.nodes)
}

// Nodes returns a slice of the nodes.
func (r Store) Nodes(props ...string) []Actioner {
	selection := ""
	filter := ""

	if len(props) > 0 {
		selection = props[0]
	}
	if len(props) > 1 {
		filter = props[1]
	}

	refs := []Actioner{}
	changed := &readOp{
		filter:    filter,
		selection: selection,
		Refs:      make(chan Actioner),
	}
	r.Read <- changed

	for ref := range changed.Refs {
		refs = append(refs, ref)
	}
	return refs
}

// ChangedRefs returns a slice of the nodes that has changed.
func (r Store) ChangedRefs() []Actioner {
	refs := []Actioner{}
	changed := &readOp{
		selection: "changed",
		Refs:      make(chan Actioner),
	}
	r.Read <- changed

	for ref := range changed.Refs {
		refs = append(refs, ref)
	}
	return refs
}

// ChangedFiles returns a slice of files that has changed.
func (r Store) ChangedFiles() []Actioner {
	refs := []Actioner{}
	changed := &readOp{
		filter:    "file",
		selection: "changed",
		Refs:      make(chan Actioner),
	}
	r.Read <- changed

	for ref := range changed.Refs {
		refs = append(refs, ref)
	}
	return refs
}

// ChangedExecs returns a slice of execs that has changed.
func (r Store) ChangedExecs() []Actioner {
	refs := []Actioner{}
	changed := &readOp{
		filter:    "exec",
		selection: "changed",
		Refs:      make(chan Actioner),
	}
	r.Read <- changed

	for ref := range changed.Refs {
		refs = append(refs, ref)
	}
	return refs
}

// ParentFiles returns a slice of all the parent files.
func (r Store) ParentFiles(file string) []string {
	parents := &readOp{
		filter:    "file",
		selection: "parents",
		node:      file,
		nodes:     make(chan string),
	}
	r.Read <- parents

	nodes := []string{}
	for node := range parents.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// ParentRefs returns a slice of all the parent refs.
func (r Store) ParentRefs(file string) []string {
	parents := &readOp{
		selection: "parents",
		node:      file,
		nodes:     make(chan string),
	}
	r.Read <- parents

	nodes := []string{}
	for node := range parents.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// Output writes the graph to a file.
func (r Store) Output() {
	r.OutputChan <- struct{}{}
}
