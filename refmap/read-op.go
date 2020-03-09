package refmap

import (
	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type changedOp struct {
	Refs chan Actioner
}

func (o changedOp) handle(refs map[string]Actioner, g *graph.Graph) {
	g.CompileRun(func(ref string) error {
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
