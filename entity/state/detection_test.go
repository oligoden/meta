package state_test

import (
	"fmt"
	"testing"

	"github.com/oligoden/meta/entity/state"
)

func TestHash(t *testing.T) {
	d := state.Detect{}

	err := d.HashOf("hash-this")
	if err != nil {
		t.Error(err)
	}

	if d.Hash() == "" {
		t.Errorf(`expected hash, got empty sting`)
	}

	exp := fmt.Sprintf("%d", state.Added)
	got := fmt.Sprintf("%d", d.State())
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	err = d.HashOf("hash-this")
	if err != nil {
		t.Error(err)
	}

	exp = status[state.Checked]
	got = status[d.State()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	err = d.HashOf("hash-something-else")
	if err != nil {
		t.Error(err)
	}

	exp = status[state.Updated]
	got = status[d.State()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

var status = map[uint8]string{
	state.Stable:  "Stable",
	state.Checked: "Checked",
	state.Updated: "Updated",
	state.Added:   "Added",
	state.Remove:  "Remove",
}
