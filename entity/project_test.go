package entity_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
	"github.com/stretchr/testify/assert"
)

func TestProjectLoadFile(t *testing.T) {
	if err := os.MkdirAll("testing", 0755); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("testing")

	c := []byte(`{"name": "abc"}`)
	if err := ioutil.WriteFile("testing/meta.json", c, 0644); err != nil {
		t.Fatal(err)
	}

	e := entity.NewProject()
	err := e.LoadFile("testing/meta.json")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("orig"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("dest"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	rm := refmap.Start()
	err = e.Process(&entity.Branch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = rm.Evaluate()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "prj:abc", e.Identifier())
}

func TestProjectProcessExt(t *testing.T) {
	f := bytes.NewBufferString(`{"name": "abc"}`)

	e := entity.NewProject()
	err := e.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	// testing overriding
	f = bytes.NewBufferString(`{"environment": "development"}`)

	err = e.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("orig"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("dest"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 0)

	rm := refmap.Start()
	err = e.Process(&entity.Branch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = rm.Evaluate()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "prj:abc", e.Identifier())
	assert.NotEmpty(t, e.Hash())
	assert.Equal(t, "development", e.Environment)
}

func TestProjectNameChange(t *testing.T) {
	assert := assert.New(t)
	f := bytes.NewBufferString(`{}`)

	e := entity.NewProject()
	err := e.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("orig"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("dest"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 3)

	rm := refmap.Start()
	err = e.Process(&entity.ProjectBranch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = rm.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	rm.Assess()
	rm.Finish()

	assert.Equal(1, len(rm.Nodes()))

	// testing reprocessing
	f = bytes.NewBufferString(`{"name": "abc"}`)

	err = e.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	err = e.Process(&entity.ProjectBranch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = rm.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	rm.Assess()
	rm.Finish()

	assert.Equal(1, len(rm.Nodes()))
	assert.Equal("prj:abc", e.Identifier())

	f = bytes.NewBufferString(`{"name": "def"}`)

	err = e.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	err = e.Process(&entity.ProjectBranch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = rm.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	rm.Assess()
	rm.Finish()

	assert.Equal(1, len(rm.Nodes()))
	assert.Equal("prj:def", e.Identifier())

	f = bytes.NewBufferString(`{"name": ""}`)

	err = e.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	err = e.Process(&entity.ProjectBranch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = rm.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	rm.Assess()
	rm.Finish()

	assert.Equal(1, len(rm.Nodes()))
	assert.Equal("prj:", e.Identifier())

	f = bytes.NewBufferString(`{"name": ""}`)

	err = e.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	err = e.Process(&entity.ProjectBranch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = rm.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	rm.Assess()
	rm.Finish()

	assert.Equal(1, len(rm.Nodes()))
	assert.Equal("prj:", e.Identifier())
}

func TestProjectReload(t *testing.T) {
	assert := assert.New(t)
	f := bytes.NewBufferString(`{
		"dirs": {
			"a": {
				"files": {
					"b":{}
				}
			}
		}
	}`)

	e := entity.NewProject()
	err := e.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("orig"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("dest"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 3)

	rm := refmap.Start()
	err = e.Process(&entity.ProjectBranch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = rm.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	rm.Assess()
	rm.Finish()

	assert.Equal(3, len(rm.Nodes()))
	assert.Equal(state.Stable, e.Directories["a"].State())
	assert.Equal(state.Stable, e.Directories["a"].Files["b"].State())
	hash := e.Directories["a"].Hash()

	// testing reprocessing
	f = bytes.NewBufferString(`{
		"directories": {
			"a": {
				"files": {
					"b":{}
				}
			}
		}
	}`)

	err = e.Load(f)
	if err != nil {
		t.Fatal(err)
	}

	err = e.Process(&entity.ProjectBranch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(hash, e.Directories["a"].Hash())
	assert.Equal(state.Checked, e.Directories["a"].State())
	assert.Equal(state.Checked, e.Directories["a"].Files["b"].State())

	err = rm.Evaluate()
	if err != nil {
		t.Fatal(err)
	}
	rm.Assess()
	rm.Finish()

	assert.Equal(3, len(rm.Nodes()))
}

func TestProjectPerform(t *testing.T) {
	if err := os.MkdirAll("testing/a", 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll("testing/b", 0755); err != nil {
		t.Fatal(err)
	}

	c := []byte("test")
	if err := ioutil.WriteFile("testing/test.ext", c, 0644); err != nil {
		t.Fatal(err)
	}

	c = []byte(`aa {{define "aa"}}aa{{end}}`)
	if err := ioutil.WriteFile("testing/a/aa.ext", c, 0644); err != nil {
		t.Fatal(err)
	}

	c = []byte(`{{template "aa"}} ab`)
	if err := ioutil.WriteFile("testing/a/ab.ext", c, 0644); err != nil {
		t.Fatal(err)
	}

	c = []byte(`{{template "aa"}} ac`)
	if err := ioutil.WriteFile("testing/a/ac.ext", c, 0644); err != nil {
		t.Fatal(err)
	}

	c = []byte(`{{template "aa"}} ba`)
	if err := ioutil.WriteFile("testing/b/ba.ext", c, 0644); err != nil {
		t.Fatal(err)
	}

	defer func() {
		os.RemoveAll("testing/test.ext")
		os.RemoveAll("testing/a")
		os.RemoveAll("testing/b")
		os.RemoveAll("testing/out")
	}()

	f := bytes.NewBufferString(`{
		"name": "abc",
		"options": "output",
		"mappings": [
			{"start": "file:a/aa.ext", "end": "file:a/ab.ext"},
			{"start": "file:a/aa.ext", "end": "file:b/ba.ext"}
		],
		"dirs": {
			"a": {
				"mappings": [
					{"start": "file:a/aa.ext", "end": "file:a/ac.ext"}
				],
				"files": {
					"aa.ext": {},
					"ab.ext": {},
					"ac.ext": {}
				}
			},
			"b": {
				"files": {
					"ba.ext": {}
				}
			}
		},
		"files": {
			"test.ext": {}
		}
	}`)

	e := entity.NewProject()
	err := e.Load(f)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}

	for _, ref := range rm.ChangedRefs() {
		err = ref.Perform(rm, ctx)
		if err != nil {
			t.Fatal(err)
		}
	}

	if _, err := os.Stat("testing/out/test.ext"); err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadFile("testing/out/test.ext")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test", string(content))

	if _, err := os.Stat("testing/out/a/aa.ext"); err != nil {
		t.Fatal(err)
	}
	content, err = ioutil.ReadFile("testing/out/a/aa.ext")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "aa ", string(content))

	if _, err := os.Stat("testing/out/a/ab.ext"); err != nil {
		t.Fatal(err)
	}
	content, err = ioutil.ReadFile("testing/out/a/ab.ext")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "aa ab", string(content))

	if _, err := os.Stat("testing/out/b/ba.ext"); err != nil {
		t.Fatal(err)
	}
	content, err = ioutil.ReadFile("testing/out/b/ba.ext")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "aa ba", string(content))
}

func TestTemplateMethodsOnProject(t *testing.T) {
	if err := os.MkdirAll("testing", 0755); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("testing")

	os.Setenv("TEST", "test")
	c := []byte(`{{.Env "TEST"}}`)
	if err := ioutil.WriteFile("testing/a.ext", c, 0644); err != nil {
		t.Error(err)
	}

	eFile := &entity.File{
		Name:   "a.ext",
		Source: "a.ext",
		Opts:   "output",
		Branch: &entity.ProjectBranch{},
	}

	eProject := &entity.Project{
		Basic: entity.Basic{
			Files: map[string]*entity.File{
				"a.ext": eFile,
			},
		},
	}

	eFile.Parent = eProject

	ctx := context.Background()
	ctx = context.WithValue(ctx, refmap.ContextKey("orig"), "testing")
	ctx = context.WithValue(ctx, refmap.ContextKey("dest"), "testing/out")
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 3)

	err := eFile.Perform(nil, ctx)
	if err != nil {
		t.Error("error performing action ->", err)
	}

	content, err := ioutil.ReadFile("testing/out/a.ext")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "test", string(content))
}
