package refmap_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/oligoden/meta/refmap"
)

func TestReadEmptyOnly(t *testing.T) {
	rm := refmap.Start("a")
	rsp := rm.Read("")
	if rsp != nil {
		fmt.Println("expected a nil response, got", rsp)
	}
}

func TestAddNewRef(t *testing.T) {
	tf1 := &testFile{
		status: refmap.DataStable,
		hash:   "a1",
	}

	rm := refmap.Start("a")
	rm.Write("b", "c", tf1)

	exp := "DataAdded"
	got := refmap.StatusText[tf1.Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	rsp := rm.Read("b")
	got = refmap.StatusText[rsp.Change]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
	if _, ok := rsp.Files["c"]; !ok {
		t.Fatal("destination not found")
	}
	got = refmap.StatusText[rsp.Files["c"].Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	rspCR := rm.ChangedRefs()
	if len(rspCR) != 1 {
		t.Error("expected changed references")
	}
	if _, ok := rspCR[0].Files["c"]; !ok {
		t.Fatal("destination not found")
	}
	exp = "a1"
	got = rspCR[0].Files["c"].Hash()
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	// adding refmap over added refmap
	tf2 := &testFile{
		status: refmap.DataStable,
		hash:   "a2",
	}
	rm.Write("b", "c", tf2)

	exp = "DataAdded"
	got = refmap.StatusText[tf2.Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	// adding refmap to existing source
	tf3 := &testFile{
		status: refmap.DataStable,
		hash:   "a3",
	}
	rm.Write("b", "d", tf3)

	exp = "DataAdded"
	got = refmap.StatusText[tf3.Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	rsp = rm.Read("b")
	if _, ok := rsp.Files["d"]; !ok {
		t.Fatal("destination not found")
	}
	got = refmap.StatusText[rsp.Files["d"].Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	rspCR = rm.ChangedRefs()
	if len(rspCR) != 1 {
		t.Error("expected changed references")
	}
	if _, ok := rspCR[0].Files["d"]; !ok {
		t.Fatal("destination not found")
	}
	exp = "a3"
	got = rspCR[0].Files["d"].Hash()
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	// // set status back to stable
	// rm.Finish()

	// exp = "DataStable"
	// got = refmap.StatusText[tf2.Change()]
	// if got != exp {
	// 	t.Errorf(`expected "%s", got "%s"`, exp, got)
	// }

	// rsp = rm.Read("a/b")
	// exp = "DataStable"
	// got = refmap.StatusText[rsp.Change]
	// if got != exp {
	// 	t.Errorf(`expected "%s", got "%s"`, exp, got)
	// }

	// // write the same file as already added
	// rm.Write("b", "c", tf2)

	// exp = "DataChecked"
	// got = refmap.StatusText[tf2.Change()]
	// if got != exp {
	// 	t.Errorf(`expected "%s", got "%s"`, exp, got)
	// }

	// rsp = rm.Read("a/b")
	// exp = "DataChecked"
	// got = refmap.StatusText[rsp.Change]
	// if got != exp {
	// 	t.Errorf(`expected "%s", got "%s"`, exp, got)
	// }

	// // set status back to stable, update file and write
	// rm.Finish()
	// tf2 = &testFile{
	// 	status: refmap.DataStable,
	// 	hash:   "a3",
	// }
	// rm.Write("b", "c", tf2)

	// exp = "DataUpdated"
	// got = refmap.StatusText[tf2.Change()]
	// if got != exp {
	// 	t.Errorf(`expected "%s", got "%s"`, exp, got)
	// }

	// rsp = rm.Read("a/b")
	// exp = "DataChecked"
	// got = refmap.StatusText[rsp.Change]
	// if got != exp {
	// 	t.Errorf(`expected "%s", got "%s"`, exp, got)
	// }

	// rspCR = rm.ChangedRefs()
	// if len(rspCR) != 1 {
	// 	t.Errorf("expected changed references, got %d", len(rspCR))
	// }
	// exp = "a1"
	// got = rspCR[0].Files["c"].Hash()
	// if got != exp {
	// 	t.Errorf(`expected "%s", got "%s"`, exp, got)
	// }
}

type testFile struct {
	status uint8
	hash   string
}

func (tf testFile) Perform(c context.Context) error {
	return nil
}

func (tf *testFile) Change(s ...uint8) uint8 {
	if len(s) > 0 {
		tf.status = s[0]
	}

	return tf.status
}

func (tf testFile) Hash() string {
	return tf.hash
}