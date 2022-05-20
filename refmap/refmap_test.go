package refmap_test

import (
	"context"
	"testing"

	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
)

func TestNormalAdding(t *testing.T) {
	rm := refmap.Start()
	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)
	t1 := &testRef{}
	rm.AddRef(ctx, "a", t1)
	t2 := &testRef{}
	rm.AddRef(ctx, "b", t2)
	t3 := &testRef{}
	rm.AddRef(ctx, "c", t3)
	rm.MapRef(ctx, "a", "b")
	rm.Evaluate()
	rm.Finish()

	exp := refmap.StatusText[state.Stable]
	got := refmap.StatusText[t1.State()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
	exp = refmap.StatusText[state.Stable]
	got = refmap.StatusText[t2.State()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	rm.SetUpdate("a")
	exp = refmap.StatusText[state.Updated]
	got = refmap.StatusText[t1.State()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	rm.Propagate()
	exp = refmap.StatusText[state.Updated]
	got = refmap.StatusText[t2.State()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestAddingUpdatedRef(t *testing.T) {
	rm := refmap.Start()
	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)
	t1 := &testRef{hash: "a", status: 3}
	rm.AddRef(ctx, "a", t1)
	t2 := &testRef{hash: "b", status: 3}
	rm.AddRef(ctx, "a", t2)
	rm.Evaluate()

	exp := "b"
	got := rm.ChangedRefs()[0].Hash()
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

type testRef struct {
	status uint8
	hash   string
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

func (tf *testRef) State(s ...uint8) uint8 {
	if len(s) > 0 {
		tf.status = s[0]
	}

	return tf.status
}

func (tf testRef) Hash() string {
	return tf.hash
}
