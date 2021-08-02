package entity_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
)

func TestLoad(t *testing.T) {
	e := &entity.Basic{}

	f := bytes.NewBufferString(`{bad json}`)
	err := e.Load(f)
	if err == nil {
		t.Error("expected error")
	}

	f = bytes.NewBufferString(`{
		"name":"abc",
		"directories":{
			"a":{}
		}
	}`)
	err = e.Load(f)
	if err != nil {
		t.Error(err)
	}

	exp := "abc"
	got := e.Name
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestBasicProcess(t *testing.T) {
	f := bytes.NewBufferString(`{
			"name":"abc"
		}`)
	e := &entity.Basic{}
	err := e.Load(f)
	if err != nil {
		t.Error(err)
	}

	branch := &entity.Branch{}
	rm := refmap.Start()

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("source"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("destination"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("watching"), false)
	ctx = context.WithValue(ctx, refmap.ContextKey("force"), true)
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	err = e.Process(branch, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	if e.Hash() == "" {
		t.Error("expected non empty hash")
	}

	err = rm.Evaluate()
	if err != nil {
		t.Error("error evaluating refmap", err)
	}

	nodes := rm.ChangedRefs()
	if len(nodes) == 0 {
		t.Fatal("expected refs in refmap")
	}

	exp := "basic:abc"
	got := fmt.Sprint(nodes[0].Identifier())
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}
