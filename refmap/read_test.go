package refmap_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/oligoden/meta/refmap"
)

func TestReadChanged(t *testing.T) {
	rm := refmap.Start()
	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	t1 := newTestRef("x")
	t1.ProcessState("x")
	rm.AddRef(ctx, "a", t1)

	t2 := newTestRef("y")
	t2.ProcessState("y")
	rm.AddRef(ctx, "b", t2)

	t3 := newTestRef("z")
	t3.ProcessState("z")
	rm.AddRef(ctx, "c", t3)

	rm.MapRef(ctx, "a", "b")
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
