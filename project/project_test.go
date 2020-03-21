package project_test

import (
	"bytes"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"

	"github.com/oligoden/meta/project"
)

func TestLoad(t *testing.T) {
	f := bytes.NewBufferString(`{bad json}`)
	_, err := project.Load(f)
	if err == nil {
		t.Error("expected error")
	}

	f = bytes.NewBufferString(`{
		"name":"abc",
		"directories":{
			"a":{}
		}
	}`)
	p, err := project.Load(f)
	if err != nil {
		t.Error(err)
	}

	exp := "abc"
	got := p.Name
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if p.Directories == nil {
		t.Error("expected non nil Directories")
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
			}
		}
	}`)
	p, err := project.Load(f)
	if err != nil {
		t.Error(err)
	}

	rm := refmap.Start()
	err = p.Process(project.BuildBranch, rm)
	if err != nil {
		t.Error(err)
	}

	exp := "a"
	got := p.Directories["a"].Name
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	exp = "a"
	got = p.Directories["a"].SourcePath
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	exp = "a"
	got = p.Directories["a"].DestinationPath
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if _, ok := p.Directories["a"].Parent.(*project.Project); !ok {
		t.Error("expected parent to be project, got")
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
	p, err := project.Load(f)
	if err != nil {
		t.Error(err)
	}

	rm := refmap.Start()
	err = p.Process(entity.BuildBranch, rm)
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
	p, err = project.Load(f)
	if err != nil {
		t.Error(err)
	}

	err = p.Process(entity.BuildBranch, rm)
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
