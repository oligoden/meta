package refmap

import "fmt"

type readOp struct {
	Src string
	Opp string
	Rsp chan *DestRef
}

func (r Store) Read(src string) *DestRef {
	read := &readOp{
		Src: src,
		Rsp: make(chan *DestRef),
	}
	r.reads <- read
	return <-read.Rsp
}

type changedOp struct {
	Refs chan *DestRef
}

func (o changedOp) handle(refs map[string]*DestRef) {
	for src, ref := range refs {
		if ref.Change == DataUpdated || ref.Change == DataAdded {
			fmt.Println("detected changed file", src)
			o.Refs <- ref
		}
	}
	close(o.Refs)
}

// ChangedRefs returns a slice of DestRefs that has changed.
func (r Store) ChangedRefs() []*DestRef {
	refs := []*DestRef{}
	for ref := range r.changedRefsChan() {
		refs = append(refs, ref)
	}
	return refs
}

func (r Store) changedRefsChan() chan *DestRef {
	changed := &changedOp{
		Refs: make(chan *DestRef),
	}
	r.Changed <- changed
	return changed.Refs
}
