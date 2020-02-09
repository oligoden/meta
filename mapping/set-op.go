package mapping

type SetOp struct {
	Key string
	Val string
	Err chan error
}

func (o SetOp) handle(location *string, refs map[string]*DestRef) {
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
	if o.Key == "finish" {
		finish(refs)
		o.Err <- nil
		return
	}

	// var value uint8
	// if o.Val == "update" {
	// 	value = DataUpdated
	// } else {
	// 	o.Err <- fmt.Errorf("unknown value")
	// 	return
	// }
	// if ref, found := refs[o.Key]; found {
	// 	ref.Change = value
	// 	fmt.Printf("setting %s status %s\n", o.Key, statusText[value])
	// 	o.Err <- nil
	// }
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

// func (r RefMap) Set(key, val string) error {
// 	setter := &SetOp{
// 		Key: key,
// 		Val: val,
// 		Err: make(chan error),
// 	}
// 	r.Sets <- setter
// 	return <-setter.Err
// }

func (r Store) Finish() {
	setter := &SetOp{
		Key: "finish",
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
