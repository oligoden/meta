package entity_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
)

func TestFileProcess(t *testing.T) {
	f := bytes.NewBufferString(`{
		"name": "abc",
		"controls": {
			"behaviour": {"options":"b,c"},
			"mappings": [
				{"start": "file:a.ext", "end": "file:b.ext"}
			]
		},
		"files": {
			"a.ext": {
				"controls": {
					"behaviour": {"options":"a,-c"}
				}
			},
			"b.ext": {}
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

	file, ok := e.Files["a.ext"]
	if !ok {
		t.Fatal("no file a.ext")
	}

	exp := "file:a.ext"
	got := file.Identifier()
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if file.Hash() == "" {
		t.Error("expected non empty hash")
	}

	exp = "&{[] a.ext {}} or &{[] b.ext {}}"
	got = fmt.Sprint(file.Branch)
	if !strings.Contains(exp, got) {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	exp = "a"
	cnt := file.Controls.Behaviour.Options
	if !strings.Contains(cnt, exp) {
		t.Errorf(`expected "%s" in "%s"`, exp, cnt)
	}

	exp = "b"
	if !strings.Contains(cnt, exp) {
		t.Errorf(`expected "%s" in "%s"`, exp, cnt)
	}

	exp = "c"
	if strings.Contains(cnt, exp) {
		t.Errorf(`did not expect "%s" in "%s"`, exp, cnt)
	}

	parents := rm.ParentFiles("file:a.ext")
	exp = "[file:a.ext]"
	got = fmt.Sprint(parents)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	parents = rm.ParentFiles("file:b.ext")
	exp = "[file:b.ext file:a.ext]"
	got = fmt.Sprint(parents)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestFilePerform(t *testing.T) {
	if err := os.MkdirAll("testing", 0755); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("testing")

	c := []byte(`a{{define "a"}}a{{end}}`)
	if err := ioutil.WriteFile("testing/a.ext", c, 0644); err != nil {
		t.Error(err)
	}

	c = []byte(`{{template "a"}} b`)
	if err := ioutil.WriteFile("testing/b.ext", c, 0644); err != nil {
		t.Error(err)
	}

	f := bytes.NewBufferString(`{
		"name": "abc",
		"controls": {
			"behaviour": {
				"options": "output"
			},
			"mappings": [
				{"start": "file:a.ext", "end": "file:b.ext"}
			]
		},
		"files": {
			"a.ext": {},
			"b.ext": {}
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

	if _, err := os.Stat("testing/out/a.ext"); err != nil {
		t.Error(err)
	}
	content, err := ioutil.ReadFile("testing/out/a.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp := "a"
	got := string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if _, err := os.Stat("testing/out/b.ext"); err != nil {
		t.Error(err)
	}
	content, err = ioutil.ReadFile("testing/out/b.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp = "a b"
	got = string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestFilePerformCopy(t *testing.T) {
	if err := os.MkdirAll("testing", 0755); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("testing")

	c := []byte(`{{"a"}}`)
	if err := ioutil.WriteFile("testing/a.ext", c, 0644); err != nil {
		t.Error(err)
	}

	f := bytes.NewBufferString(`{
		"name": "abc",
		"controls": {
			"behaviour": {
				"options": "output,copy"
			}
		},
		"files": {
			"a.ext": {}
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

	if _, err := os.Stat("testing/out/a.ext"); err != nil {
		t.Error(err)
	}
	content, err := ioutil.ReadFile("testing/out/a.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp := `{{"a"}}`
	got := string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestFilters(t *testing.T) {
	if err := os.MkdirAll("testing", 0755); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("testing")

	c := []byte(`//-a`)
	if err := ioutil.WriteFile("testing/a.ext", c, 0644); err != nil {
		t.Error(err)
	}

	f := bytes.NewBufferString(`{
		"name": "abc",
		"controls": {
			"behaviour": {
				"options": "output",
				"filters": {"comment":{}}
			}
		},
		"files": {
			"a.ext": {}
		}
	}`)

	e := &entity.Basic{}
	err := e.Load(f)
	if err != nil {
		t.Error("error loading config", err)
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

	if _, err := os.Stat("testing/out/a.ext"); err != nil {
		t.Error(err)
	}
	content, err := ioutil.ReadFile("testing/out/a.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp := ""
	got := string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}
