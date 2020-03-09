package refmap

import (
	"fmt"

	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type addOp struct {
	key string
	val Actioner
	rsp chan error
}

func (o addOp) handle(refs map[string]Actioner, g *graph.Graph) {
	if _, found := refs[o.key]; !found {
		refs[o.key] = o.val
		refs[o.key].State(state.Added)
		g.Add(o.key)
	} else {
		if refs[o.key].State() == state.Added {
			refs[o.key] = o.val
			refs[o.key].State(state.Added)
			fmt.Printf("WARNING - duplicate %s added\n", o.key)
		} else {
			if refs[o.key].Hash() == o.val.Hash() {
				o.val.State(state.Checked)
			} else {
				refs[o.key] = o.val
				refs[o.key].State(state.Updated)
			}
		}
	}

	o.rsp <- nil
}

func (r Store) AddRef(key string, val Actioner) {
	add := &addOp{
		key: key,
		val: val,
		rsp: make(chan error),
	}
	r.Adds <- add
	<-add.rsp
	return
}

type mapOp struct {
	start string
	end   string
	set   uint
	rsp   chan error
}

func (r Store) MapRef(key0, key1 string, setOption ...uint) {
	set := uint(1)
	if len(setOption) > 0 {
		set = setOption[0]
	}

	m := &mapOp{
		start: key0,
		end:   key1,
		set:   set,
		rsp:   make(chan error),
	}
	r.Maps <- m
	<-m.rsp
	return
}
