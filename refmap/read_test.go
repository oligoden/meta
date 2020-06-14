package refmap_test

import (
	"fmt"
	"testing"

	"github.com/oligoden/meta/refmap"
)

func TestReadChanged(t *testing.T) {
	rm := refmap.Start()
	t1 := &testRef{}
	rm.AddRef("a", t1)
	t2 := &testRef{}
	rm.AddRef("b", t2)
	t3 := &testRef{}
	rm.AddRef("c", t3)
	rm.MapRef("a", "b")
	rm.Evaluate()

	rsp := rm.ChangedRefs()
	if len(rsp) != 3 {
		fmt.Println("expected 3 changed refs, got", len(rsp))
	}

	rspNodes := rm.ParentFiles("b")
	if len(rspNodes) != 2 {
		fmt.Println("expected 2 changed refs, got", len(rspNodes))
	}
}
