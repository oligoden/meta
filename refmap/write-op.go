package refmap

import (
	"fmt"

	graph "github.com/oligoden/math-graph"
)

type writeOp struct {
	src string
	dst string
	val Actioner
	rsp chan bool
}

func (o writeOp) handle(location string, refs map[string]*DestRef) {
	if _, found := refs[o.src]; !found {
		refs[o.src] = NewDestination()
		fmt.Printf("adding source %s\n", o.src)
	} else {
		if refs[o.src].Change == DataStable {
			refs[o.src].Change = DataChecked
		}
	}

	if _, found := refs[o.src].Files[o.dst]; !found {
		refs[o.src].Files[o.dst] = o.val
		refs[o.src].Files[o.dst].Change(DataAdded)
		fmt.Printf("adding destination refmap %s -> %s\n", o.src, o.dst)
	} else {
		if refs[o.src].Files[o.dst].Change() == DataAdded {
			o.val.Change(DataAdded)
			fmt.Printf("WARNING - duplicate %s -> %s added, over-writing previous entry\n", o.src, o.dst)
		} else {
			if refs[o.src].Files[o.dst].Hash() == o.val.Hash() {
				o.val.Change(DataChecked)
			} else {
				o.val.Change(DataUpdated)
				fmt.Printf("updating %s -> %s to status %s\n", o.src, o.dst, StatusText[refs[o.src].Files[o.dst].Change()])
			}
		}
		refs[o.src].Files[o.dst] = o.val
	}

	o.rsp <- true
}

func (r Store) Write(src, dst string, val Actioner) {
	write := &writeOp{
		src: src,
		dst: dst,
		val: val,
		rsp: make(chan bool),
	}
	r.writes <- write
	<-write.rsp
	return
}

type addOp struct {
	key string
	val Actioner
	rsp chan error
}

func (o addOp) handle(refs map[string]Actioner, g *graph.Graph) {
	if _, found := refs[o.key]; !found {
		refs[o.key] = o.val
		refs[o.key].Change(DataAdded)
		g.Add(o.key)
	} else {
		if refs[o.key].Change() == DataAdded {
			refs[o.key] = o.val
			refs[o.key].Change(DataAdded)
			fmt.Printf("WARNING - duplicate %s added\n", o.key)
		} else {
			if refs[o.key].Hash() == o.val.Hash() {
				o.val.Change(DataChecked)
			} else {
				refs[o.key] = o.val
				refs[o.key].Change(DataUpdated)
			}
		}
	}

	o.rsp <- nil
}

func (r Store) AddRef(kind, path string, val Actioner) {
	key := fmt.Sprintf("%s-%s", kind, path)
	add := &addOp{
		key: key,
		val: val,
		rsp: make(chan error),
	}
	r.adds <- add
	<-add.rsp
	return
}

type mapOp struct {
	start string
	end   string
	set   uint
	rsp   chan error
}

func (r Store) MapRef(kind0, path0, kind1, path1 string, setOption ...uint) {
	key0 := fmt.Sprintf("%s-%s", kind0, path0)
	key1 := fmt.Sprintf("%s-%s", kind1, path1)
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
