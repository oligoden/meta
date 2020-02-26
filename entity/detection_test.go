package entity_test

import (
	"fmt"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
)

func TestHash(t *testing.T) {
	d := entity.Detection{}

	err := d.HashOf("hash-this")
	if err != nil {
		t.Error(err)
	}

	if d.Hash() == "" {
		t.Errorf(`expected hash, got empty sting`)
	}

	exp := fmt.Sprintf("%d", refmap.DataAdded)
	got := fmt.Sprintf("%d", d.Change())
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	err = d.HashOf("hash-this")
	if err != nil {
		t.Error(err)
	}

	exp = status[refmap.DataChecked]
	got = status[d.Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	err = d.HashOf("hash-something-else")
	if err != nil {
		t.Error(err)
	}

	exp = status[refmap.DataUpdated]
	got = status[d.Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

var status = map[uint8]string{
	refmap.DataStable:  "DataStable",
	refmap.DataChecked: "DataChecked",
	refmap.DataUpdated: "DataUpdated",
	refmap.DataAdded:   "DataAdded",
	refmap.DataRemove:  "DataRemove",
}
