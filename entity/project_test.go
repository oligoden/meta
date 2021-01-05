package entity_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
)

func TestLoad(t *testing.T) {
	f := bytes.NewBufferString(`{bad json}`)
	_, err := entity.Load(f)
	if err == nil {
		t.Error("expected error")
	}

	f = bytes.NewBufferString(`{
		"name":"abc",
		"directories":{
			"a":{}
		}
	}`)
	p, err := entity.Load(f)
	if err != nil {
		t.Error(err)
	}

	exp := "abc"
	got := p.Name
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if p.Directories == nil {
		t.Error("expected directories")
	}
}

func TestLoadOverride(t *testing.T) {
	f := bytes.NewBufferString(`{
			"name":"abc",
			"execs": {
				"a": {
					"cmd": ["a"]
				},
				"b": {}
			}
		}`)
	p, err := entity.Load(f)
	if err != nil {
		t.Error(err)
	}

	f = bytes.NewBufferString(`{bad json}`)
	err = p.Load(f)
	if err == nil {
		t.Error("expected error")
	}

	f = bytes.NewBufferString(`{
			"name":"def",
			"execs": {
				"a": {
					"cmd": ["b"]
				},
				"c": {}
			}
		}`)

	err = p.Load(f)
	if err != nil {
		t.Error(err)
	}

	exp := "def"
	got := p.Name
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if p.Execs == nil {
		t.Error("expected execs")
		t.FailNow()
	}

	if len(p.Execs) != 3 {
		t.Error("expected 3 execs")
	}

	exc, ok := p.Execs["a"]
	if !ok {
		t.Error("expected exec a")
		t.FailNow()
	}

	if len(exc.Cmd) != 1 {
		t.Error("expected 1 cmd arg")
	}
}

func TestProcess(t *testing.T) {
	f := bytes.NewBufferString(`{
		"name":"abc",
		"directories":{
			"a":{
				"files":{
					"aa.ext":{}
				}
			},
			"b":{
				"files":{
					"ba.ext":{}
				}
			}
		}
	}`)
	p, err := entity.Load(f)
	if err != nil {
		t.Error(err)
	}

	rm := refmap.Start()
	ctx := context.WithValue(context.Background(), entity.ContextKey("verbose"), 0)
	err = p.Process(&entity.ProjectBranch{}, rm, ctx)
	if err != nil {
		t.Error(err)
	}

	exp := "a"
	got := p.Directories["a"].Name
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	exp = "a"
	got = p.Directories["a"].SrcDerived
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	exp = "a"
	got = p.Directories["a"].DstDerived
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if _, ok := p.Directories["a"].Parent.(*entity.Project); !ok {
		t.Error("expected parent to be project")
	}

	if p.Directories["a"].Hash() == "" {
		t.Errorf(`expected hash, got empty sting`)
	}
}

func TestProcessCheckHashChange(t *testing.T) {
	f := bytes.NewBufferString(`{
			"name":"abc",
			"directories":{
				"a":{}
			}
		}`)
	p, err := entity.Load(f)
	if err != nil {
		t.Error(err)
	}

	rm := refmap.Start()
	ctx := context.WithValue(context.Background(), entity.ContextKey("verbose"), 0)
	err = p.Process(&entity.ProjectBranch{}, rm, ctx)
	if err != nil {
		t.Error(err)
	}

	hash1 := p.Basic.Hash()

	f = bytes.NewBufferString(`{
			"name":"abc",
			"directories":{
				"a":{},
				"b":{}
			}
		}`)
	p, err = entity.Load(f)
	if err != nil {
		t.Error(err)
	}

	err = p.Process(&entity.ProjectBranch{}, rm, ctx)
	if err != nil {
		t.Error(err)
	}

	hash2 := p.Basic.Hash()

	if hash1 == "" {
		t.Errorf(`expected hash, got empty sting`)
	}

	if hash2 != hash1 {
		t.Errorf(`expected hash to stay constant`)
	}
}
