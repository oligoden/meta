package refmap

import (
	"context"
	"fmt"

	graph "github.com/oligoden/math-graph"
)

type addOp struct {
	key string
	val Actioner
	ctx context.Context
	rsp chan error
}

func (o addOp) handle(refs map[string]Actioner, g *graph.Graph) {
	if _, found := refs[o.key]; !found {
		verboseValue := o.ctx.Value(ContextKey("verbose")).(int)
		if verboseValue >= 3 {
			fmt.Println("adding", o.key)
		}
		refs[o.key] = o.val
		g.Add(o.key)
	} else {
		fmt.Println("already got", o.key)
		refs[o.key] = o.val
	}

	o.rsp <- nil
}

func (r Store) AddRef(ctx context.Context, key string, val Actioner) {
	add := &addOp{
		key: key,
		val: val,
		ctx: ctx,
		rsp: make(chan error),
	}
	r.Adds <- add
	<-add.rsp
}

type rnmOp struct {
	key string
	val string
	rsp chan error
}

func (o rnmOp) handle(refs map[string]Actioner, g *graph.Graph) {
	if _, found := refs[o.key]; !found {
		o.rsp <- fmt.Errorf("ref %s does not exist", o.key)
		return
	}
	refs[o.key] = refs[o.val]
	delete(refs, o.key)
	g.Rename(o.key, o.val)

	o.rsp <- nil
}

func (r Store) RenameRef(ctx context.Context, key string, val string) {
	verboseValue := ctx.Value(ContextKey("verbose")).(int)
	if verboseValue >= 3 {
		fmt.Println("renaming", key, "to", val)
	}

	rn := &rnmOp{
		key: key,
		val: val,
		rsp: make(chan error),
	}
	r.Rnms <- rn
	<-rn.rsp
}

type mapOp struct {
	start string
	end   string
	set   uint
	rsp   chan error
}

func (r Store) MapRef(ctx context.Context, key0, key1 string, setOption ...uint) error {
	verboseValue := ctx.Value(ContextKey("verbose")).(int)
	if verboseValue >= 3 {
		fmt.Println("mapping", key0, key1)
	}

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
