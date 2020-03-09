package refmap

import (
	"fmt"

	graph "github.com/oligoden/math-graph"
	"github.com/oligoden/meta/entity/state"
)

type SetOp struct {
	Key string
	Val string
	Err chan error
}

func (o SetOp) handle(refs map[string]Actioner, g *graph.Graph) {
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
		propagate(refs, g)
		o.Err <- nil
	case "evaluate":
		g.Evaluate()
		o.Err <- nil
	case "finish":
		// finish(refs)
		finish(refs)
		o.Err <- nil
	default:
		var value uint8
		if o.Val == "update" {
			value = state.Updated
		} else {
			o.Err <- fmt.Errorf("unknown value")
			return
		}
		if ref, found := refs[o.Key]; found {
			ref.State(value)
			o.Err <- nil
			return
		}
		o.Err <- fmt.Errorf("key not found")
	}

}

// func assess(refs map[string]*RefLink) {
// 	for _, ref := range refs {
// 		if ref.Change == state.Stable {
// 			ref.Change = state.Remove
// 			continue
// 		}
// 		for _, file := range ref.Files {
// 			if file.Change() == state.Stable {
// 				file.Change(state.Remove)
// 			}
// 		}
// 	}
// }

// func finish(refs map[string]*DestRef) {
// 	for _, ref := range refs {
// 		// 		if ref.Change == state.Remove {
// 		// 			fmt.Println("removing", src, "from refmap")
// 		// 			delete(refs, src)
// 		// 			continue
// 		// 		}

// 		ref.Change = state.Stable
// 		// 		if ref.Change != state.Flagged && ref.Change != state.Stable {
// 		// 			fmt.Printf("setting %s status %s\n", src, statusText[ref.Change])
// 		// 		}

// 		for _, file := range ref.Files {
// 			// 			if file.Change() == state.Remove {
// 			// 				fmt.Printf("removing %s -> %s from refmap\n", src, dst)
// 			// 				delete(ref.Files, dst)
// 			// 				continue
// 			// 			}
// 			// 			if file.Change() != state.Flagged && file.Change() != state.Stable {
// 			// 				fmt.Printf("setting %s -> %s status %s\n", src, dst, statusText[file.Change()])
// 			// 			}
// 			file.State(state.Stable)
// 		}
// 	}
// }

func propagate(refs map[string]Actioner, g *graph.Graph) {
	for node := range g.StartNodes() {
		update := false
		g.SetRun(func(node string) error {
			if refs[node].State() == state.Updated {
				update = true
			}
			if update {
				refs[node].State(state.Updated)
			}
			return nil
		}, node)
	}
}

func finish(refs map[string]Actioner) {
	for _, ref := range refs {
		// 		if ref.Change == state.Remove {
		// 			fmt.Println("removing", src, "from refmap")
		// 			delete(refs, src)
		// 			continue
		// 		}

		ref.State(state.Stable)
	}
}

func (r Store) SetUpdate(key string) error {
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
