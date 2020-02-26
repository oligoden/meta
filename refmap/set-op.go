package refmap

import (
	"fmt"

	graph "github.com/oligoden/math-graph"
)

type SetOp struct {
	Key string
	Val string
	Err chan error
}

func (o SetOp) handle(refs map[string]*DestRef, refs2 map[string]Actioner, g *graph.Graph) {
	// if o.Key == "location" {
	// 	*location = o.Val
	// 	o.Err <- nil
	// 	return
	// }
	// if o.Key == "assess" {
	// 	assess(refs)
	// 	o.Err <- nil
	// 	return
	// }
	switch o.Key {
	case "propagate":
		propagate(refs2, g)
		o.Err <- nil
	case "evaluate":
		g.Evaluate()
		fmt.Printf("%+v", g)
		o.Err <- nil
	case "finish":
		finish(refs)
		finish2(refs2)
		o.Err <- nil
	default:
		var value uint8
		if o.Val == "update" {
			value = DataUpdated
		} else {
			o.Err <- fmt.Errorf("unknown value")
			return
		}
		if ref, found := refs2[o.Key]; found {
			ref.Change(value)
			o.Err <- nil
			return
		}
		o.Err <- fmt.Errorf("key not found")
	}

}

// func assess(refs map[string]*RefLink) {
// 	for _, ref := range refs {
// 		if ref.Change == DataStable {
// 			ref.Change = DataRemove
// 			continue
// 		}
// 		for _, file := range ref.Files {
// 			if file.Change() == DataStable {
// 				file.Change(DataRemove)
// 			}
// 		}
// 	}
// }

func finish(refs map[string]*DestRef) {
	for _, ref := range refs {
		// 		if ref.Change == DataRemove {
		// 			fmt.Println("removing", src, "from refmap")
		// 			delete(refs, src)
		// 			continue
		// 		}

		ref.Change = DataStable
		// 		if ref.Change != DataFlagged && ref.Change != DataStable {
		// 			fmt.Printf("setting %s status %s\n", src, statusText[ref.Change])
		// 		}

		for _, file := range ref.Files {
			// 			if file.Change() == DataRemove {
			// 				fmt.Printf("removing %s -> %s from refmap\n", src, dst)
			// 				delete(ref.Files, dst)
			// 				continue
			// 			}
			// 			if file.Change() != DataFlagged && file.Change() != DataStable {
			// 				fmt.Printf("setting %s -> %s status %s\n", src, dst, statusText[file.Change()])
			// 			}
			file.Change(DataStable)
		}
	}
}

func propagate(refs map[string]Actioner, g *graph.Graph) {
	for node := range g.StartNodes() {
		update := false
		g.SetRun(func(node string) error {
			if refs[node].Change() == DataUpdated {
				update = true
			}
			if update {
				refs[node].Change(DataUpdated)
			}
			return nil
		}, node)
	}
}

func finish2(refs map[string]Actioner) {
	for _, ref := range refs {
		// 		if ref.Change == DataRemove {
		// 			fmt.Println("removing", src, "from refmap")
		// 			delete(refs, src)
		// 			continue
		// 		}

		ref.Change(DataStable)
	}
}

func (r Store) SetUpdate(kind, path string) error {
	key := fmt.Sprintf("%s-%s", kind, path)

	setter := &SetOp{
		Key: key,
		Val: "update",
		Err: make(chan error),
	}
	r.Sets <- setter
	return <-setter.Err
}

func (r Store) Finish() {
	setter := &SetOp{
		Key: "finish",
		Err: make(chan error),
	}
	r.Sets <- setter
	<-setter.Err
}

func (r Store) Evaluate() {
	setter := &SetOp{
		Key: "evaluate",
		Err: make(chan error),
	}
	r.Sets <- setter
	<-setter.Err
}

func (r Store) Propagate() {
	setter := &SetOp{
		Key: "propagate",
		Err: make(chan error),
	}
	r.Sets <- setter
	<-setter.Err
}

// func (r RefMap) Assess() {
// 	setter := &SetOp{
// 		Key: "assess",
// 		Err: make(chan error),
// 	}
// 	r.Sets <- setter
// 	<-setter.Err
// }
