//go:build linux || darwin
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
	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
	"github.com/stretchr/testify/assert"
)

func TestExecProcess(t *testing.T) {
	f := bytes.NewBufferString(`{
		"name": "abc",
		"mappings": [
			{"start": "file:a.ext", "end": "exec:cp"}
		],
		"files": {
			"a.ext": {}
		},
		"execs": {
			"cp": {
				"cmd": ["cp", "a.ext", "b.ext"],
				"env": {"E":"a"}
			}
		}
	}`)

	e := &entity.Basic{Detect: state.New()}
	err := e.Load(f)
	if err != nil {
		t.Error("loading config")
	}

	rm := refmap.Start()

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("orig"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("dest"), "testing/out")
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

	if assert.Contains(t, exec.Env, "E") {
		assert.Equal(t, "a", exec.Env["E"])
	}

	parents := rm.ParentFiles("exec:cp")
	exp = "[file:a.ext]"
	got = fmt.Sprint(parents)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestExecPerform(t *testing.T) {
	assert := assert.New(t)

	if err := os.MkdirAll("testing/cd", 0755); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("testing")

	c := []byte(`a`)
	if err := ioutil.WriteFile("testing/cd/a.ext", c, 0644); err != nil {
		t.Error(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("orig"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("dest"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	cle := &entity.CLE{}
	cle.Name = "cp"
	cle.Cmd = []string{"cp", "a.ext", "b.ext"}
	cle.Env = map[string]string{"E": "a"}
	cle.Dir = "cd"

	assert.NoError(cle.Perform(nil, ctx))

	if assert.NotNil(cle) {
		assert.Equal("a", cle.Env["E"])
		assert.Equal("exec:cp", cle.Identifier())
		assert.Equal("action cp was run", cle.Output())
		assert.Equal("cd", cle.Dir)
	}

	content, err := ioutil.ReadFile("testing/cd/b.ext")
	if assert.NoError(err) {
		assert.Equal("a", string(content))
	}
}
