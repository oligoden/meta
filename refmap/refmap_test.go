package refmap_test

import (
	"context"
	"testing"

	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
	"github.com/stretchr/testify/assert"
)

func TestNormalAdding(t *testing.T) {
	assert := assert.New(t)

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
	rm.Finish()

	assert.Equal(state.Stable, t1.State())
	assert.Equal(state.Stable, t2.State())

	rm.SetUpdate("a")
	assert.Equal(state.Updated, t1.State())

	rm.Propagate()
	assert.Equal(state.Updated, t2.State())
}

func TestAddingUpdatedRef(t *testing.T) {
	rm := refmap.Start()
	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	t1 := newTestRef("x")
	t1.ProcessState("x")
	rm.AddRef(ctx, "a", t1)

	t2 := newTestRef("y")
	t2.ProcessState("y")
	hash := t2.Hash()
	rm.AddRef(ctx, "a", t2)

	rm.Evaluate()

	assert.Equal(t, hash, rm.ChangedRefs()[0].Hash())
}

func TestRemoving(t *testing.T) {
	assert := assert.New(t)

	rm := refmap.Start()
	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	t1 := newTestRef("x")
	t1.ProcessState("x")
	rm.AddRef(ctx, "a", t1)

	t2 := newTestRef("y")
	t2.ProcessState("y")
	rm.AddRef(ctx, "b", t2)

	rm.Evaluate()
	rm.Finish()

	assert.Len(rm.Nodes(), 2)

	t1.ProcessState("x")
	rm.Assess()
	assert.Equal(state.Checked, t1.State())
	assert.Equal(state.Remove, t2.State())
	rm.Finish()

	assert.Len(rm.Nodes(), 1)
}

type testRef struct {
	Name string
	*state.Detect
}

func newTestRef(name string) *testRef {
	return &testRef{
		Name:   name,
		Detect: state.New(),
	}
}

func (testRef) Identifier() string {
	return ""
}

func (testRef) Output() string {
	return ""
}

func (testRef) Perform(rm refmap.Grapher, c context.Context) error {
	return nil
}
