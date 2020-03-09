package refmap_test

import (
	"context"
	"testing"

	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
)

func TestNormalAdding(t *testing.T) {
	rm := refmap.Start()
	t1 := &testRef{}
	rm.AddRef("a", t1)
	t2 := &testRef{}
	rm.AddRef("b", t2)
	t3 := &testRef{}
	rm.AddRef("c", t3)
	rm.MapRef("a", "b")
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

func TestAddingRefOverExistingRef(t *testing.T) {
	rm := refmap.Start()
	t1 := &testRef{}
	rm.AddRef("a", t1)
	t2 := &testRef{}
	rm.AddRef("a", t2)
	rm.Finish()

	exp := refmap.StatusText[state.Added]
	got := refmap.StatusText[t1.State()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
	exp = refmap.StatusText[state.Stable]
	got = refmap.StatusText[t2.State()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestAddingSameRef(t *testing.T) {
	rm := refmap.Start()
	t1 := &testRef{hash: "abc"}
	rm.AddRef("a", t1)
	rm.Finish()

	rm.AddRef("a", t1)
	exp := refmap.StatusText[state.Checked]
	got := refmap.StatusText[t1.State()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestAddingUpdatedRef(t *testing.T) {
	rm := refmap.Start()
	t1 := &testRef{hash: "abc"}
	rm.AddRef("a", t1)
	rm.Finish()

	t2 := &testRef{hash: "def"}
	rm.AddRef("a", t2)
	exp := refmap.StatusText[state.Updated]
	got := refmap.StatusText[t2.State()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

type testRef struct {
	status uint8
	hash   string
}

func (testRef) Perform(c context.Context) error {
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
