package refmap_test

import (
	"context"
	"testing"

	"github.com/oligoden/meta/refmap"
)

func TestSetting(t *testing.T) {
	rm := refmap.Start("")
	t1 := &testRef{}
	rm.AddRef("k", "a", t1)
	t2 := &testRef{}
	rm.AddRef("k", "b", t2)
	t3 := &testRef{}
	rm.AddRef("k", "c", t3)
	rm.MapRef("k", "a", "k", "b")
	rm.Evaluate()

	rm.Finish()
	exp := refmap.StatusText[refmap.DataStable]
	got := refmap.StatusText[t1.Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
	exp = refmap.StatusText[refmap.DataStable]
	got = refmap.StatusText[t2.Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	rm.SetUpdate("k", "a")
	exp = refmap.StatusText[refmap.DataUpdated]
	got = refmap.StatusText[t1.Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	rm.Propagate()
	exp = refmap.StatusText[refmap.DataUpdated]
	got = refmap.StatusText[t2.Change()]
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

func (tf *testRef) Change(s ...uint8) uint8 {
	if len(s) > 0 {
		tf.status = s[0]
	}

	return tf.status
}

func (tf testRef) Hash() string {
	return tf.hash
}
