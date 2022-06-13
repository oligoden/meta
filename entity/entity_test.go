package entity_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	assert := assert.New(t)
	e := &entity.Basic{}

	f := bytes.NewBufferString(`{bad json}`)
	assert.Error(e.Load(f))

	f = bytes.NewBufferString(`{
		"name":"abc"
	}`)
	assert.NoError(e.Load(f))

	assert.Equal("abc", e.Name)
}

func TestBasicProcess(t *testing.T) {
	assert := assert.New(t)
	e := &entity.Basic{
		Name:   "abc",
		Detect: state.New(),
	}

	branch := &entity.Branch{}
	rm := refmap.Start()
	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("orig"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("dest"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	if assert.NoError(e.Process(branch, rm, ctx)) {
		assert.NotEmpty(e.Hash())
	}

	err := rm.Evaluate()
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

func TestImportProcess(t *testing.T) {
	if err := os.MkdirAll("testing", 0755); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("testing")

	c := []byte(`{
		"dirs":{"a":{}},
		"files":{"a.ext":{}}
	}`)
	if err := ioutil.WriteFile("testing/meta.json", c, 0644); err != nil {
		t.Fatal(err)
	}

	f := bytes.NewBufferString(`{
		"name":"abc",
		"import":true
	}`)
	e := &entity.Basic{Detect: state.New()}
	err := e.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	branch := &entity.Branch{}
	rm := refmap.Start()

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("orig"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("dest"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	err = e.Process(branch, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, e.Files)
	assert.NotEmpty(t, e.Directories)
}
