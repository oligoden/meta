// +build linux darwin

package entity_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
)

func TestExecProcess(t *testing.T) {
	f := bytes.NewBufferString(`{
		"name": "abc",
		"controls": {
			"mappings": [
				{"start": "file:a.ext", "end": "exec:cp"}
			]
		},
		"files": {
			"a.ext": {}
		},
		"execs": {
			"cp": {
				"cmd": ["cp", "a.ext", "b.ext"]
			}
		}
	}`)

	e := &entity.Basic{}
	err := e.Load(f)
	if err != nil {
		t.Error("loading config")
	}

	rm := refmap.Start()

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("source"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("destination"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	err = e.Process(&entity.Branch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = rm.Evaluate()
	if err != nil {
		t.Error("error evaluating refmap", err)
	}

	exec, ok := e.Execs["cp"]
	if !ok {
		t.Fatal(`no exec "cp"`)
	}

	exp := "exec:cp"
	got := exec.Identifier()
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if exec.Hash() == "" {
		t.Error("expected non empty hash")
	}

	parents := rm.ParentFiles("exec:cp")
	exp = "[file:a.ext]"
	got = fmt.Sprint(parents)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestExecPerform(t *testing.T) {
	c := []byte(`a{{define "a"}}a{{end}}`)
	if err := ioutil.WriteFile("testing/a.ext", c, 0644); err != nil {
		t.Error(err)
	}

	defer func() {
		os.RemoveAll("testing/a.ext")
		os.RemoveAll("testing/out")
	}()

	f := bytes.NewBufferString(`{
		"name": "abc",
		"controls": {
			"behaviour": {
				"options": "output"
			},
			"mappings": [
				{"start": "file:a.ext", "end": "exec:cp"}
			]
		},
		"files": {
			"a.ext": {}
		},
		"execs": {
			"cp": {
				"cmd": ["cp", "a.ext", "b.ext"]
			}
		}
	}`)

	e := &entity.Basic{}
	err := e.Load(f)
	if err != nil {
		t.Error("loading config")
	}

	rm := refmap.Start()

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("source"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("destination"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	err = e.Process(&entity.Branch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = rm.Evaluate()
	if err != nil {
		t.Error("error evaluating refmap", err)
	}

	for _, ref := range rm.ChangedRefs() {
		err = ref.Perform(rm, ctx)
		if err != nil {
			t.Error("error performing action ->", err)
		}
	}

	if _, err := os.Stat("testing/out/b.ext"); err != nil {
		t.Error(err)
	}
	content, err := ioutil.ReadFile("testing/out/b.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp := "a"
	got := string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}
