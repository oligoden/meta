package refmap

import (
	graph "github.com/oligoden/math-graph"
)

type addOp struct {
	key string
	val Actioner
	rsp chan error
}

func (o addOp) handle(refs map[string]Actioner, g *graph.Graph) {
	if _, found := refs[o.key]; !found {
		refs[o.key] = o.val
		g.Add(o.key)
	} else {
		refs[o.key] = o.val
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

func (r Store) MapRef(key0, key1 string, setOption ...uint) error {
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
	return <-m.rsp
}
