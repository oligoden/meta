package entity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
)

func TestDirProcess(t *testing.T) {
	f := bytes.NewBufferString(`{
		"name": "abc",
		"controls": {
			"behaviour": {"options": "a"},
			"mappings": [
				{"start": "file:aa.ext", "end": "file:ac.ext"},
				{"start": "file:aa.ext", "end": "file:b/ba.ext"},
				{"start": "file:aa.ext", "end": "file:a.ext"}
			]
		},
		"directories": {
			"a": {
				"controls": {
					"mappings": [
						{"start": "file:aa.ext", "end": "file:ab.ext"}
					]
				},
				"src-ovr": "/",
				"dst-ovr": "aa",
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
			"a.ext": {}
		}
	}`)

	e := &entity.Basic{}
	err := e.Load(f)
	if err != nil {
		t.Fatal("error loading config ->", err)
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

	dir, ok := e.Directories["a"]
	if !ok {
		t.Fatal("no dir a")
	}

	exp := "a"
	got := dir.Name
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if dir.Hash() == "" {
		t.Error("expected non empty hash")
	}

	exp = ""
	got = dir.SrcDerived
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	exp = "a/aa"
	got = dir.DstDerived
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	exp = "a"
	got = dir.Options()
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	parents := rm.ParentFiles("file:ab.ext")
	exp = "[file:ab.ext file:aa.ext]"
	got = fmt.Sprint(parents)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	parents = rm.ParentFiles("file:ac.ext")
	exp = "[file:ac.ext file:aa.ext]"
	got = fmt.Sprint(parents)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	parents = rm.ParentFiles("file:b/ba.ext")
	exp = "[file:b/ba.ext file:aa.ext]"
	got = fmt.Sprint(parents)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	parents = rm.ParentFiles("file:a.ext")
	exp = "[file:a.ext file:aa.ext]"
	got = fmt.Sprint(parents)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestDirPerform(t *testing.T) {
	if err := os.MkdirAll("testing", 0755); err != nil {
		t.Error(err)
	}
	defer os.RemoveAll("testing")

	c := []byte(`a {{template "aa"}}`)
	if err := ioutil.WriteFile("testing/a.ext", c, 0644); err != nil {
		t.Error(err)
	}

	c = []byte(`{{define "aa"}}aa{{end}}`)
	if err := ioutil.WriteFile("testing/aa.ext", c, 0644); err != nil {
		t.Error(err)
	}

	c = []byte(`ab {{template "aa"}}`)
	if err := ioutil.WriteFile("testing/ab.ext", c, 0644); err != nil {
		t.Error(err)
	}

	c = []byte(`ac {{template "aa"}}`)
	if err := ioutil.WriteFile("testing/ac.ext", c, 0644); err != nil {
		t.Error(err)
	}

	if err := os.MkdirAll("testing/aa", 0755); err != nil {
		t.Error(err)
	}

	c = []byte(`aaa {{template "aa"}}`)
	if err := ioutil.WriteFile("testing/aa/aaa.ext", c, 0644); err != nil {
		t.Error(err)
	}

	if err := os.MkdirAll("testing/b", 0755); err != nil {
		t.Error(err)
	}

	c = []byte(`ba {{template "aa"}}`)
	if err := ioutil.WriteFile("testing/b/ba.ext", c, 0644); err != nil {
		t.Error(err)
	}

	f := bytes.NewBufferString(`{
		"name": "abc",
		"controls": {
			"behaviour": {"options":"output"},
			"mappings": [
				{"start": "file:aa.ext", "end": "file:ac.ext"},
				{"start": "file:aa.ext", "end": "file:b/ba.ext"},
				{"start": "file:aa.ext", "end": "file:a.ext"}
			]
		},
		"directories": {
			"a": {
				"controls": {
					"mappings": [
						{"start": "file:aa.ext", "end": "file:ab.ext"},
						{"start": "file:aa.ext", "end": "file:aa/aaa.ext"}
					]
				},
				"src-ovr": "/",
				"files": {
					"aa.ext": {},
					"ab.ext": {},
					"ac.ext": {}
				},
				"directories": {
					"aa": {
						"files": {
							"aaa.ext": {}
						}
					}
				}
			},
			"b": {
				"files": {
					"ba.ext": {}
				}
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
	ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), 3)

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
	exp := "a aa"
	got := string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if _, err := os.Stat("testing/out/a/ac.ext"); err != nil {
		t.Error(err)
	}
	content, err = ioutil.ReadFile("testing/out/a/ac.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp = "ac aa"
	got = string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if _, err := os.Stat("testing/out/a/ab.ext"); err != nil {
		t.Error(err)
	}
	content, err = ioutil.ReadFile("testing/out/a/ab.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp = "ab aa"
	got = string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if _, err := os.Stat("testing/out/a/aa/aaa.ext"); err != nil {
		t.Error(err)
	}
	content, err = ioutil.ReadFile("testing/out/a/aa/aaa.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp = "aaa aa"
	got = string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	if _, err := os.Stat("testing/out/b/ba.ext"); err != nil {
		t.Error(err)
	}
	content, err = ioutil.ReadFile("testing/out/b/ba.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp = "ba aa"
	got = string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestSimpleDirStructure(t *testing.T) {
	t.SkipNow()
	cfg := `{
		"directories": {
			"a": {
				"directories": {
					"aa": {}
				},
				"files": {
					"aa.ext": {}
				}
			}
		}
	}`

	testCases := []struct {
		desc        string
		srcOverride string
		dstOverride string
		srcExp      string
		dstExp      string
	}{
		{
			desc:   "normal behaviour",
			srcExp: "a",
			dstExp: "a",
		},
		{
			desc:        "add to source path",
			srcExp:      "a/b",
			dstExp:      "a",
			srcOverride: "b",
		},
		{
			desc:        "destination stay in current dir",
			srcExp:      "a",
			dstExp:      ".",
			dstOverride: ".",
		},
		{
			desc:        "source goto dir in current dir",
			srcExp:      "b",
			dstExp:      "a",
			srcOverride: "./b",
		},
		{
			desc:        "destination goto root",
			srcExp:      "a",
			dstExp:      "",
			dstOverride: "/",
		},
		{
			desc:        "source goto dir in root",
			srcExp:      "b",
			dstExp:      "a",
			srcOverride: "/b",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mc := dirTestConfigSetup(cfg, t)
			dir := mc.Directories["a"]
			dir.SrcOverride = tC.srcOverride
			dir.DstOverride = tC.dstOverride
			ctx := context.Background()
			dirTestProcess(mc, ctx, t)

			// test derived paths
			exp := tC.srcExp
			got := dir.SrcDerived
			if got != exp {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}

			exp = tC.dstExp
			got = dir.DstDerived
			if got != exp {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}
		})
	}
}

func TestDeepDirStructure(t *testing.T) {
	t.SkipNow()
	cfg := `{
		"directories": {
			"a": {
				"directories": {
					"aa": {
						"directories": {
							"aaa": {}
						},
						"files": {
							"aaa.ext": {}
						}
					}
				}
			}
		}
	}`

	testCases := []struct {
		desc        string
		srcOverride string
		dstOverride string
		srcExp      string
		dstExp      string
	}{
		{
			desc:        "destination stay in current dir",
			srcExp:      "a/aa",
			dstExp:      "a",
			dstOverride: ".",
		},
		{
			desc:        "source goto dir in current dir",
			srcExp:      "a/bb",
			dstExp:      "a/aa",
			srcOverride: "./bb",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mc := dirTestConfigSetup(cfg, t)
			dir := mc.Directories["a"]
			dirAA := dir.Directories["aa"]
			dirAA.SrcOverride = tC.srcOverride
			dirAA.DstOverride = tC.dstOverride
			ctx := context.Background()
			dirTestProcess(mc, ctx, t)

			// test derived paths
			exp := tC.srcExp
			got := dirAA.SrcDerived
			if got != exp {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}

			exp = tC.dstExp
			got = dirAA.DstDerived
			if got != exp {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}
		})
	}
}

// func TestDirImport(t *testing.T) {
// 	cfg := `{
// 		"directories": {
// 			"b": {
// 				"import": {}
// 			}
// 		}
// 	}`

// 	mc := dirTestConfigSetup(cfg, t)
// 	dir := mc.Directories["b"]
// 	dir.DstDerived = "b"
// 	dir.SrcDerived = "b"
// 	dir.Name = "b"
// 	dir.Parent = mc
// 	ctx := context.Background()
// 	ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/work")
// 	dirTestProcess(mc, ctx, t)

// 	// dirA, ok := dir.Directories["ba"]
// 	// if !ok {
// 	// 	t.Fatalf(`expected to find dir "ba"`)
// 	// }
// 	// exp := "b/ba/bab.ext"
// 	// got := dirA.Files["baa.ext"].Templates[0]
// 	// if got != exp {
// 	// 	t.Errorf(`expected "%s", got "%s"`, exp, got)
// 	// }
// }

func TestDirParams(t *testing.T) {
	t.SkipNow()
	cfg := `{
		"directories": {
			"a": {
				"directories": {
					"aa": {}
				},
				"files": {
					"aa.ext": {}
				}
			}
		}
	}`

	mc := dirTestConfigSetup(cfg, t)
	dir := mc.Directories["a"]
	ctx := context.Background()
	dirTestProcess(mc, ctx, t)

	// test hash
	if dir.Hash() == "" {
		t.Errorf("expected hash, got empty string")
	}

	// test parent
	dirAA := dir.Directories["aa"]
	if d, ok := dirAA.Parent.(*entity.Directory); !ok {
		t.Error("parent is not a directory")
	} else {
		exp := dir.Hash()
		got := d.Hash()
		if got != exp {
			t.Errorf(`parent does not match, expected %+v, got %+v`, dir, d)
		}
	}

	// test name
	exp := "aa"
	got := dirAA.Name
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	// test parent ID
	exp = "dir:" + dir.SrcDerived + ":" + dir.Name
	got = dirAA.Parent.Identifier()
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}
}

func TestDirFileParams(t *testing.T) {
	t.SkipNow()
	cfg := `{
		"directories": {
			"a": {
				"directories": {
					"aa": {}
				},
				"files": {
					"aa.ext": {}
				}
			}
		}
	}`

	mc := dirTestConfigSetup(cfg, t)
	dir := mc.Directories["a"]
	ctx := context.Background()
	dirTestProcess(mc, ctx, t)

	file := dir.Files["aa.ext"]

	if file.Hash() == "" {
		t.Errorf("expected hash, got empty string")
	}

	exp := "aa.ext"
	got := file.Name
	if got != exp {
		t.Errorf("expected '%s', got '%s'", exp, got)
	}

	if d, ok := file.Parent.(*entity.Directory); !ok {
		t.Error("parent is not a directory")
	} else {
		exp := dir.Hash()
		got := d.Hash()
		if got != exp {
			t.Errorf(`parent does not match, expected %+v, got %+v`, dir, d)
		}
	}

	exp = "dir:" + dir.SrcDerived + ":" + dir.Name
	got = file.Parent.Identifier()
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	exp = filepath.Join(dir.SrcDerived, "aa.ext")
	got = file.Source
	if got != exp {
		t.Errorf("expected '%s', got '%s'", exp, got)
	}

	exp = "file:a/aa.ext"
	got = file.Identifier()
	if got != exp {
		t.Errorf("expected '%s', got '%s'", exp, got)
	}
}

func TestDirFileBranch(t *testing.T) {
	t.SkipNow()
	cfg := `{
		"directories": {
			"a": {
				"files": {
					"aa.ext": {}
				}
			}
		}
	}`

	mc := dirTestConfigSetup(cfg, t)
	dir := mc.Directories["a"]
	ctx := context.Background()
	dirTestProcess(mc, ctx, t)

	b, ok := dir.Files["aa.ext"].Branch.(*entity.Branch)
	if !ok {
		t.Fatalf("expected branch to be of type *Branch")
	}
	exp := "aa.ext"
	got := b.Filename
	if got != exp {
		t.Errorf("expected '%s', got '%s'", exp, got)
	}
}

func TestDirFileBehaviourParams(t *testing.T) {
	t.SkipNow()
	cfg := `{
		"directories": {
			"a": {
				"controls": {
					%s
				},
				"directories": {
					"aa": {}
				},
				"files": {
					"aa.ext": {}
				}
			}
		}
	}`

	testCases := []struct {
		desc          string
		ctrl          string
		dirBhOptions  string
		dirBhOutput   bool
		dirBhFilter   map[string]map[string]string
		fileBhOptions string
		fileBhOutput  bool
		fileBhFilter  map[string]map[string]string
	}{
		{
			desc: "none",
			ctrl: "",
			// dirBhFilter: "[]",
		},
		{
			desc: "normal behaviour",
			ctrl: `"behaviour":{}`,
			// dirBhFilter: "[]",
		},
		{
			desc:         "copy behaviour",
			ctrl:         `"behaviour":{"options":"copy"}`,
			dirBhOptions: entity.CopyBehaviour,
			// dirBhFilter:  "[]",
			fileBhOptions: entity.CopyBehaviour,
		},
		{
			desc:        "output behaviour",
			ctrl:        `"behaviour":{"output":true}`,
			dirBhOutput: true,
			// dirBhFilter:  "[]",
			fileBhOutput: true,
		},
		{
			desc:         "filter behaviour",
			ctrl:         `"behaviour":{"filters":{"comment-filter":{}}}`,
			dirBhFilter:  map[string]map[string]string{"comment-filter": {}},
			fileBhFilter: map[string]map[string]string{"comment-filter": {}},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			icfg := fmt.Sprintf(cfg, tC.ctrl)
			mc := dirTestConfigSetup(icfg, t)
			ctx := context.Background()
			dirTestProcess(mc, ctx, t)

			dir := mc.Directories["a"]
			dirAA := dir.Directories["aa"]
			fileAA := dir.Files["aa.ext"]

			if dirAA.Controls.Behaviour == nil {
				t.Fatal("expected behaviour")
			}

			exp := tC.dirBhOptions
			got := dirAA.Controls.Behaviour.Options
			if got != exp {
				t.Errorf("expected '%s', got '%s'", exp, got)
			}

			// exp = fmt.Sprintf("%v", tC.dirBhOutput)
			// got = fmt.Sprintf("%v", dirAA.Controls.Behaviour.Output)
			// if got != exp {
			// 	t.Errorf("expected '%s', got '%s'", exp, got)
			// }

			exp = fmt.Sprintf("%v", tC.dirBhFilter)
			got = fmt.Sprintf("%v", dirAA.Controls.Behaviour.Filters)
			if got != exp {
				t.Errorf("expected '%s', got '%s'", exp, got)
			}

			if fileAA.Controls.Behaviour == nil {
				t.Fatal("expected behaviour")
			}

			if fileAA.Controls.Behaviour == nil {
				t.Fatal("expected behaviour")
			}
			exp = tC.fileBhOptions
			got = fileAA.Controls.Behaviour.Options
			if got != exp {
				t.Errorf("expected '%s', got '%s'", exp, got)
			}
			// exp = fmt.Sprintf("%v", tC.fileBhOutput)
			// got = fmt.Sprintf("%v", fileAA.Controls.Behaviour.Output)
			// if got != exp {
			// 	t.Errorf("expected '%s', got '%s'", exp, got)
			// }
			exp = fmt.Sprintf("%v", tC.fileBhFilter)
			got = fmt.Sprintf("%v", fileAA.Controls.Behaviour.Filters)
			if got != exp {
				t.Errorf("expected '%s', got '%s'", exp, got)
			}
		})
	}
}

func TestDirRefMappings(t *testing.T) {
	t.SkipNow()
	cfg := `{
		"directories": {
			"a": {
				"files": {
					"aa.ext": {}
				},
				"execs": {
					"aa": {}
				}
			}
		}
	}`

	mc := dirTestConfigSetup(cfg, t)
	dir := mc.Directories["a"]
	ctx := context.Background()
	rm := dirTestProcess(mc, ctx, t)

	expDir := "dir:" + dir.SrcDerived + ":" + dir.Name
	_, fnd := rm.nodes[expDir]
	if !fnd {
		t.Fatalf(`expected to find "%s"`, expDir)
	}
	_, fnd = rm.maps[expDir]
	if !fnd {
		t.Fatalf(`expected to find "%s"`, expDir)
	}

	expFile := "file:" + filepath.Join(dir.SrcDerived, "aa.ext")
	_, fnd = rm.nodes[expFile]
	if !fnd {
		t.Fatalf(`expected to find "%s"`, expFile)
	}
	lnk, fnd := rm.maps[expDir][expFile]
	if !fnd {
		t.Fatalf(`expected to find "%s"`, expFile)
	}
	if !lnk {
		t.Errorf(`expected to see that "%s" and "%s" is linked`, expDir, expFile)
	}

	expExec := "exec:aa"
	_, fnd = rm.nodes[expExec]
	if !fnd {
		t.Fatalf(`expected to find "%s"`, expExec)
	}
	lnk, fnd = rm.maps[expDir][expExec]
	if !fnd {
		t.Fatalf(`expected to find "%s"`, expExec)
	}
	if !lnk {
		t.Errorf(`expected to see that "%s" and "%s" is linked`, expDir, expExec)
	}
}

func TestDirMappingParams(t *testing.T) {
	t.SkipNow()
	cfg := `{
		"directories": {
			"a": {
				"controls": {
					"mappings": [
						{
							"start": "file:a\/aa.ext",
							"end": "file:a\/ab.ext"
						},
						{
							"start": "file:a\/aa\/aaa.ext",
							"end": "file:a\/ab\/aba.ext",
							"recurrence": 1
						},
						{
							"start": "file:a\/aa\/aaa.ext",
							"end": "file:a\/aa.ext",
							"recurrence": 1
						},
						{
							"start": "exec:aaa$",
							"end": "exec:aa$",
							"recurrence": 1
						}
					]
				},
				"directories": {
					"aa": {
						"controls": {
							"mappings": [
								{
									"start": "file:a\/aa\/aaa.ext",
									"end": "exec:aaa"
								}
							]
						},
						"files": {
							"aaa.ext": {}
						},
						"execs": {
							"aaa": {}
						}
					},
					"ab": {
						"files": {
							"aba.ext": {}
						}
					}
				},
				"files": {
					"aa.ext": {},
					"ab.ext": {}
				},
				"execs": {
					"aa": {}
				}
			}
		}
	}`

	mc := dirTestConfigSetup(cfg, t)
	mc.Name = "project:name"
	ctx := context.Background()
	rm := dirTestProcess(mc, ctx, t)

	start, ok := rm.maps["file:a/aa/aaa.ext"]
	if !ok {
		t.Fatal("expected start node file:a/aa/aaa.ext")
	}
	_, ok = start["file:a/aa.ext"]
	if !ok {
		t.Fatal("expected end node file:a/aa.ext")
	}

	start, ok = rm.maps["file:a/aa/aaa.ext"]
	if !ok {
		t.Fatal("expected start node file:a/aa/aaa.ext")
	}
	_, ok = start["file:a/ab/aba.ext"]
	if !ok {
		t.Fatal("expected end node file:a/ab/aba.ext")
	}

	start, ok = rm.maps["exec:aaa"]
	if !ok {
		t.Fatal("expected start node exec:aaa")
	}
	_, ok = start["exec:aa"]
	if !ok {
		t.Fatal("expected end node exec:aa")
	}

	start, ok = rm.maps["file:a/aa.ext"]
	if !ok {
		t.Fatal("expected start node file:a/aa.ext")
	}
	_, ok = start["file:a/ab.ext"]
	if !ok {
		t.Fatal("expected end node file:a/ab.ext")
	}

	start, ok = rm.maps["file:a/aa/aaa.ext"]
	if !ok {
		t.Fatal("expected start node file:a/aa/aaa.ext")
	}
	_, ok = start["exec:aaa"]
	if !ok {
		t.Fatal("expected end node exec:aaa")
	}
}

func TestDirExecParams(t *testing.T) {
	t.SkipNow()
	cfg := `{
		"directories": {
			"a": {
				"directories": {
					"aa": {}
				},
				"execs": {
					"aa": {}
				}
			}
		}
	}`

	mc := dirTestConfigSetup(cfg, t)
	dir := mc.Directories["a"]
	ctx := context.Background()
	dirTestProcess(mc, ctx, t)

	exec := dir.Execs["aa"]

	exp := "aa"
	got := exec.Name
	if got != exp {
		t.Errorf("expected '%s', got '%s'", exp, got)
	}

	if d, ok := exec.Parent.(*entity.Directory); !ok {
		t.Error("parent is not a directory")
	} else {
		exp := dir.Hash()
		got := d.Hash()
		if got != exp {
			t.Errorf(`parent does not match, expected %+v, got %+v`, dir, d)
		}
	}

	if exec.Hash() == "" {
		t.Errorf("expected hash, got empty string")
	}
}

// func TestDirectoryReprocessingSame(t *testing.T) {
// 	str := `{
// 		"directories": {
// 			"a": {
// 				"directories": {
// 					"aa": {

// 					}
// 				},
// 				"files": {
// 					"aa.ext": {

// 					}
// 				}
// 			}
// 		}
// 		}`

// 	rm := refMapStub{
// 		nodes: map[string]refmap.Actioner{},
// 		maps:  map[string]map[string]bool{},
// 	}
// 	m := entity.Basic{}
// 	json.Unmarshal([]byte(str), &m)
// 	m.Directories["a"].DestinationPath = "a"
// 	m.Directories["a"].SourcePath = "a"
// 	m.Directories["a"].Name = "a"
// 	m.Directories["a"].ParentID = "project:name"
// 	m.Directories["a"].Parent = m
// 	ctx := context.WithValue(context.Background(), entity.ContextKey("verbose"), 0)
// 	m.Directories["a"].Process(entity.BuildBranch, rm, ctx)
// 	hash := m.Directories["a"].Hash()

// 	m.Directories["a"].Files["aa.ext"].Settings = "copy-only"
// 	m.Directories["a"].Directories["aa"].Settings = "copy-only"
// 	m.Directories["a"].Process(entity.BuildBranch, rm, ctx)

// 	if m.Directories["a"].Hash() != hash {
// 		t.Error("expected hash to stay the same")
// 	}
// }

// func TestDirectoryReprocessingUpdate(t *testing.T) {
// 	str := `{
// 		"directories": {
// 			"a": {
// 				"directories": {
// 					"aa": {

// 					}
// 				},
// 				"files": {
// 					"aa.ext": {

// 					}
// 				}
// 			}
// 		}
// 		}`

// 	rm := refMapStub{
// 		nodes: map[string]refmap.Actioner{},
// 		maps:  map[string]map[string]bool{},
// 	}
// 	m := entity.Basic{}
// 	json.Unmarshal([]byte(str), &m)
// 	m.Directories["a"].DestinationPath = "a"
// 	m.Directories["a"].SourcePath = "a"
// 	m.Directories["a"].Name = "a"
// 	m.Directories["a"].ParentID = "project:name"
// 	m.Directories["a"].Parent = m
// 	ctx := context.WithValue(context.Background(), entity.ContextKey("verbose"), 0)
// 	m.Directories["a"].Process(entity.BuildBranch, rm, ctx)
// 	hash := m.Directories["a"].Hash()

// 	m.Directories["a"].Settings = "copy-only"
// 	m.Directories["a"].Process(entity.BuildBranch, rm, ctx)

// 	if m.Directories["a"].Hash() == hash {
// 		t.Error("expected hash to change")
// 	}
// }

// func TestDirectoryReprocessingLoadUpdate(t *testing.T) {
// 	str := `{
// 		"directories": {
// 			"a": {
// 				"directories": {
// 					"aa": {

// 					}
// 				},
// 				"files": {
// 					"aa.ext": {

// 					}
// 				}
// 			}
// 		}
// 	}`

// 	rm := refMapStub{
// 		nodes: map[string]refmap.Actioner{},
// 		maps:  map[string]map[string]bool{},
// 	}
// 	m := entity.Basic{}
// 	json.Unmarshal([]byte(str), &m)
// 	m.Directories["a"].DestinationPath = "a"
// 	m.Directories["a"].SourcePath = "a"
// 	m.Directories["a"].Name = "a"
// 	m.Directories["a"].ParentID = "project:name"
// 	m.Directories["a"].Parent = m

// 	ctx := context.WithValue(context.Background(), entity.ContextKey("verbose"), 0)
// 	m.Directories["a"].Process(entity.BuildBranch, rm, ctx)
// 	hash := m.Directories["a"].Hash()

// 	str = `{
// 		"directories": {
// 			"a": {
// 				"copy-only": true,
// 				"directories": {
// 					"aa": {

// 					}
// 				},
// 				"files": {
// 					"aa.ext": {

// 					}
// 				}
// 			}
// 		}
// 	}`
// 	json.Unmarshal([]byte(str), &m)
// 	m.Directories["a"].Process(entity.BuildBranch, rm, ctx)

// 	if m.Directories["a"].Hash() == hash {
// 		t.Error("expected hash to change")
// 	}
// }

func dirTestConfigSetup(cfg string, t *testing.T) *entity.Basic {
	// mc symbolises the meta config struct
	mc := &entity.Basic{}
	err := json.Unmarshal([]byte(cfg), mc)
	if err != nil {
		t.Fatal(err)
	}

	return mc
}

func dirTestProcess(mc *entity.Basic, ctx context.Context, t *testing.T) refMapStub {
	// these are set in project processing
	var dir *entity.Directory
	for k := range mc.Directories {
		dir = mc.Directories[k]
		dir.DstDerived = k
		dir.SrcDerived = k
		dir.Name = k
		dir.Parent = mc
		break
	}

	rm := refMapStub{
		nodes: map[string]refmap.Actioner{},
		maps:  map[string]map[string]bool{},
	}

	err := dir.Process(&entity.Branch{}, rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	return rm
}

type refMapStub struct {
	nodes map[string]refmap.Actioner
	maps  map[string]map[string]bool
}

func (rm refMapStub) AddRef(ctx context.Context, d string, f refmap.Actioner) {
	rm.nodes[d] = f
}

func (rm refMapStub) MapRef(ctx context.Context, a, b string, o ...uint) error {
	if rm.maps[a] == nil {
		rm.maps[a] = map[string]bool{
			b: true,
		}
		return nil
	}
	rm.maps[a][b] = true
	return nil
}

func (rm refMapStub) ParentFiles(f string) []string {
	return []string{}
}
