package refmap

import (
	"strings"

	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type changedOp struct {
	kind string
	Refs chan Actioner
}

func (o changedOp) handle(refs map[string]Actioner, g *graph.Graph) {
	g.CompileRun(func(ref string) error {
		if !strings.HasPrefix(ref, o.kind) {
			return nil
		}
		if refs[ref].State() == state.Updated || refs[ref].State() == state.Added {
			o.Refs <- refs[ref]
		}
		return nil
	})
	close(o.Refs)
}

// ChangedRefs returns a slice of DestRefs that has changed.
func (r Store) ChangedRefs() []Actioner {
	refs := []Actioner{}
	changed := &changedOp{
		Refs: make(chan Actioner),
	}
	r.Changed <- changed

	for ref := range changed.Refs {
		refs = append(refs, ref)
	}
	return refs
}

// ChangedFiles returns a slice of files that has changed.
func (r Store) ChangedFiles() []Actioner {
	refs := []Actioner{}
	changed := &changedOp{
		kind: "file",
		Refs: make(chan Actioner),
	}
	r.Changed <- changed

	for ref := range changed.Refs {
		refs = append(refs, ref)
	}
	return refs
}

// ChangedExecs returns a slice of execs that has changed.
func (r Store) ChangedExecs() []Actioner {
	refs := []Actioner{}
	changed := &changedOp{
		kind: "exec",
		Refs: make(chan Actioner),
	}
	r.Changed <- changed

	for ref := range changed.Refs {
		refs = append(refs, ref)
	}
	return refs
}
