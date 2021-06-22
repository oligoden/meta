package entity_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
)

func TestFile(t *testing.T) {
	// d1 := []byte("add this\nthis should not be here")
	// ioutil.WriteFile("testing/a/aa/aaa.go", d1, 0644)

	testCases := []struct {
		desc        string
		file        string
		parentCtrls string
		fileCtrls   string
		output      bool
		content     string
	}{
		{
			desc:   "default no output behaviour",
			file:   "aa.ext",
			output: false,
		},
		{
			desc:      "output set on file",
			file:      "aa.ext",
			content:   "abc",
			fileCtrls: `"controls":{"behaviour":{"options":"output"}}`,
			output:    true,
		},
		{
			desc:        "output set on parent",
			file:        "aa.ext",
			content:     "abc",
			parentCtrls: `"controls":{"behaviour":{"options":"output"}},`,
			output:      true,
		},
	}

	for _, tC := range testCases {
		dstFilename := strings.TrimSuffix(tC.file, ".tmpl")

		t.Run(tC.desc, func(t *testing.T) {
			str := `{
				%s
				"files": {
					"%s": {%s}
				}
			}`
			str = fmt.Sprintf(str, tC.parentCtrls, tC.file, tC.fileCtrls)

			n := &entity.Basic{}
			err := json.Unmarshal([]byte(str), &n)
			if err != nil {
				t.Error("error unmarshalling,", err)
			}
			n.Name = "node:name"

			rm := refmap.Start()
			rm.AddRef("node:name", n)

			ctx := context.Background()
			ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/work")
			ctx = context.WithValue(ctx, entity.ContextKey("destination"), "testing")
			ctx = context.WithValue(ctx, entity.ContextKey("watching"), false)
			ctx = context.WithValue(ctx, entity.ContextKey("force"), true)
			ctx = context.WithValue(ctx, entity.ContextKey("verbose"), 0)

			err = n.Process(&entity.Branch{}, rm, ctx)
			if err != nil {
				t.Fatal(err)
			}

			file, ok := n.Files[tC.file]
			if !ok {
				t.Fatalf(`no file "%s"`, tC.file)
			}
			// fmt.Printf("%+v\n", n)
			fmt.Printf("file %+v\n", file)
			fmt.Printf("file.Controls %+v\n", file.Controls)
			fmt.Printf("file.Controls.Behaviour %+v\n\n", file.Controls.Behaviour)

			exp := "file:" + tC.file
			got := file.Identifier()
			if exp != got {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}

			if file.Hash() == "" {
				t.Error("expected non empty hash")
			}

			if file.OptionsContain("output") != tC.output {
				t.Fatal("output flag does not match")
			}

			err = file.Perform(rm, ctx)
			if err != nil {
				t.Fatal(err)
			}

			if file.OptionsContain("output") {
				if _, err := os.Stat("testing/" + dstFilename); err != nil {
					t.Error(err)
				}

				content, err := ioutil.ReadFile("testing/" + dstFilename)
				if err != nil {
					t.Fatal(err)
				}
				exp := tC.content
				got := string(content)
				if exp != got {
					t.Errorf(`expected "%s", got "%s"`, exp, got)
				}
			}
		})

		os.RemoveAll("testing/" + dstFilename)
	}
}

// func TestFileForcing(t *testing.T) {
// 	str := `{
// 		"directories": {
// 			"a": {
// 				"directories": {
// 					"aa": {
// 						"files": {
// 							"aab.ext": {}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}`
// 	str = fmt.Sprintf(str)

// 	m := &entity.Basic{}
// 	err := json.Unmarshal([]byte(str), &m)
// 	if err != nil {
// 		t.Error("error unmarshalling,", err)
// 	}

// 	dir := m.Directories["a"]
// 	dir.DstDerived = "a"
// 	dir.SrcDerived = "a"
// 	dir.Name = "a"
// 	// m.Directories["a"].ParentID = "project:name"
// 	m.Directories["a"].Parent = m

// 	rm := refmap.Start()
// 	rm.AddRef("project:name", m)

// 	ctx := context.Background()
// 	ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/work")
// 	ctx = context.WithValue(ctx, entity.ContextKey("destination"), "testing")
// 	ctx = context.WithValue(ctx, entity.ContextKey("watching"), false)
// 	ctx = context.WithValue(ctx, entity.ContextKey("force"), true)
// 	ctx = context.WithValue(ctx, entity.ContextKey("verbose"), 0)

// 	err = dir.Process(&entity.Branch{}, rm, ctx)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	f1, _ := os.Create("testing/a/aa/aab.ext")
// 	f1.WriteString("abc")
// 	f1.Close()

// 	f1, _ = os.Create("testing/work/a/aa/aab.ext")
// 	f1.WriteString("def")
// 	f1.Close()

// 	err = dir.Directories["aa"].Files["aab.ext"].Perform(rm, ctx)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	content, err := ioutil.ReadFile("testing/a/aa/aab.ext")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	exp := "def"
// 	got := string(content)
// 	if exp != got {
// 		t.Errorf(`expected "%s", got "%s"`, exp, got)
// 	}

// 	os.RemoveAll("testing/work/a/aa/aab.ext")
// 	os.RemoveAll("testing/a/aa/aab.ext")
// }

// func TestFileNotForcing(t *testing.T) {
// 	str := `{
// 		"directories": {
// 			"a": {
// 				"directories": {
// 					"aa": {
// 						"files": {
// 							"aab.ext": {}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}`
// 	str = fmt.Sprintf(str)

// 	m := &entity.Basic{}
// 	err := json.Unmarshal([]byte(str), &m)
// 	if err != nil {
// 		t.Error("error unmarshalling,", err)
// 	}

// 	dir := m.Directories["a"]
// 	dir.DstDerived = "a"
// 	dir.SrcDerived = "a"
// 	dir.Name = "a"
// 	// m.Directories["a"].ParentID = "project:name"
// 	m.Directories["a"].Parent = m

// 	rm := refmap.Start()
// 	rm.AddRef("project:name", m)

// 	ctx := context.Background()
// 	ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/work")
// 	ctx = context.WithValue(ctx, entity.ContextKey("destination"), "testing")
// 	ctx = context.WithValue(ctx, entity.ContextKey("watching"), false)
// 	ctx = context.WithValue(ctx, entity.ContextKey("force"), false)
// 	ctx = context.WithValue(ctx, entity.ContextKey("verbose"), 0)

// 	err = dir.Process(&entity.Branch{}, rm, ctx)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	f1, _ := os.Create("testing/a/aa/aab.ext")
// 	f1.WriteString("abc")
// 	f1.Close()

// 	f1, _ = os.Create("testing/work/a/aa/aab.ext")
// 	f1.WriteString("def")
// 	f1.Close()

// 	err = dir.Directories["aa"].Files["aab.ext"].Perform(rm, ctx)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	content, err := ioutil.ReadFile("testing/a/aa/aab.ext")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	exp := "abc"
// 	got := string(content)
// 	if exp != got {
// 		t.Errorf(`expected "%s", got "%s"`, exp, got)
// 	}

// 	os.RemoveAll("testing/work/a/aa/aab.ext")
// 	os.RemoveAll("testing/a/aa/aab.ext")
// }
