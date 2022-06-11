package refmap_test

import (
	"context"
	"testing"

	"github.com/oligoden/meta/refmap"
	"github.com/stretchr/testify/assert"
)

func TestReadChanged(t *testing.T) {
	assert := assert.New(t)

	rm := refmap.Start()
	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	t1 := newTestRef("x")
	t1.ProcessState("x")
	rm.AddRef(ctx, "file", t1)
	rm.Evaluate()
	assert.Len(rm.Nodes(), 1)
	assert.Len(rm.ChangedFiles(), 1)

	t2 := newTestRef("y")
	t2.ProcessState("y")
	rm.AddRef(ctx, "exec", t2)
	rm.Evaluate()
	assert.Len(rm.Nodes(), 2)
	assert.Len(rm.ChangedExecs(), 1)

	t3 := newTestRef("z")
	t3.ProcessState("z")
	rm.AddRef(ctx, "c", t3)

	rm.MapRef(ctx, "a", "b")
	rm.Evaluate()
	assert.Len(rm.Nodes(), 3)
	assert.Len(rm.ChangedRefs(), 3)

	rm.Finish()
	assert.Len(rm.Nodes(), 3)

	// rspNodes := rm.ParentFiles("b")
	// if len(rspNodes) != 2 {
	// 	fmt.Println("expected 2 changed refs, got", len(rspNodes))
	// }
}
