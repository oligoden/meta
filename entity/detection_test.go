package entity_test

import (
	"fmt"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/mapping"
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

	exp := fmt.Sprintf("%d", mapping.DataAdded)
	got := fmt.Sprintf("%d", d.Change())
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	err = d.HashOf("hash-this")
	if err != nil {
		t.Error(err)
	}

	exp = status[mapping.DataChecked]
	got = status[d.Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	err = d.HashOf("hash-something-else")
	if err != nil {
		t.Error(err)
	}

	exp = status[mapping.DataUpdated]
	got = status[d.Change()]
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

var status = map[uint8]string{
	mapping.DataStable:  "DataStable",
	mapping.DataChecked: "DataChecked",
	mapping.DataUpdated: "DataUpdated",
	mapping.DataAdded:   "DataAdded",
	mapping.DataRemove:  "DataRemove",
}
