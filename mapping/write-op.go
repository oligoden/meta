package mapping

import (
	"fmt"
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
		fmt.Printf("adding destination mapping %s -> %s\n", o.src, o.dst)
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

// // Mutator writes a file mapping and Actioner to the RefMap.
// type Mutator interface {
// 	Write(string, string, Actioner)
// }

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
