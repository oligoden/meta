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

func TestFilePerforming(t *testing.T) {
	d1 := []byte("add this\nthis should not be here")
	ioutil.WriteFile("testing/a/aa/aaa.go", d1, 0644)

	testCases := []struct {
		desc           string
		file           string
		parentDirCtrls string
		dirCtrls       string
		fileCtrls      string
		output         bool
		content        string
	}{
		{
			desc:    "default no output behaviour",
			file:    "aaa.ext",
			content: "abc aaa.ext",
			output:  false,
		},
		// {
		// 	desc:      "template execution output",
		// 	file:      "aaa.ext",
		// 	content:   "abc aaa.ext",
		// 	fileCtrls: `"controls":{"behaviour":{"options":"output"}}`,
		// 	output:    true,
		// },
		// {
		// 	desc:    "test source property",
		// 	file:    "aaa.ext",
		// 	prps:    `"source":"./aaz.ext"`,
		// 	content: "def",
		// 	output:  true,
		// },
		// {
		// 	desc:    "test source in sub directory",
		// 	file:    "aaa.ext",
		// 	prps:    `"source":"./sub/aax.ext"`,
		// 	content: "ijk",
		// 	output:  true,
		// },
		// {
		// 	desc:    "test source in parent directory",
		// 	file:    "aab.ext",
		// 	prps:    `"source":"aa.ext"`,
		// 	content: "abc",
		// 	output:  true,
		// },
		// {
		// 	desc:    "test removal of .tmpl",
		// 	file:    "aaa.ext.tmpl",
		// 	content: "ghi",
		// 	output:  true,
		// },
		// {
		// 	desc:    "test copy only set on file",
		// 	file:    "aaa.ext",
		// 	prps:    `"settings":"copy-only"`,
		// 	content: "abc {{.Filename}}",
		// },
		// {
		// 	desc:    "test templates on file",
		// 	file:    "aa-comp.ext",
		// 	prps:    `"templates":["a/aa/aa-incl.ext"]`,
		// 	content: "yul gar jom",
		// 	output:  true,
		// },
		// {
		// 	desc:    "test copy only set on directory",
		// 	file:    "aaa.ext",
		// 	dirControls: `copy-only`,
		// 	content: "abc {{.Filename}}",
		// },
		// {
		// 	desc:         "test copy only set on directory",
		// 	file:         "aaa.ext",
		// 	fileControls: `"controls":{"behaviour":{"options":"copy","output":true}},`,
		// 	content:      "abc {{.Filename}}",
		// 	output:       true,
		// },
		// {
		// 	desc:         "test no output set on directory",
		// 	file:         "aaa.ext",
		// 	fileControls: `"controls":{"behaviour":{"output":false}},`,
		// 	output:       false,
		// },
		// {
		// 	desc: "test line inclusion control of .go files",
		// 	file: "aaa.go.tmpl",
		// 	// prps:    `"settings":"comment-filter"`,
		// 	fileControls: `"controls":{"behaviour":{"filters":{"comment-filter":{}}, "output":true}},`,
		// 	content:      "add this\n",
		// 	output:       true,
		// },
	}

	for _, tC := range testCases {
		dstFilename := strings.TrimSuffix("a/aa/"+tC.file, ".tmpl")

		t.Run(tC.desc, func(t *testing.T) {
			str := `{
				"directories": {
					"a": {
						%s
						"directories": {
							"aa": {
								%s
								"files": {
									"%s": {%s}
								}
							}
						}
					}
				}
			}`
			str = fmt.Sprintf(str, tC.parentDirCtrls, tC.dirCtrls, tC.file, tC.fileCtrls)

			m := &entity.Basic{}
			err := json.Unmarshal([]byte(str), &m)
			if err != nil {
				t.Error("error unmarshalling,", err)
			}
			m.Name = "project:name"

			dir := m.Directories["a"]
			dir.Parent = m
			dir.DstDerived = "a"
			dir.SrcDerived = "a"
			dir.Name = "a"

			rm := refmap.Start()
			rm.AddRef("project:name", m)

			ctx := context.Background()
			ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/work")
			ctx = context.WithValue(ctx, entity.ContextKey("destination"), "testing")
			ctx = context.WithValue(ctx, entity.ContextKey("watching"), false)
			ctx = context.WithValue(ctx, entity.ContextKey("force"), true)
			ctx = context.WithValue(ctx, entity.ContextKey("verbose"), 0)

			err = dir.Process(&entity.Branch{}, rm, ctx)
			if err != nil {
				t.Fatal(err)
			}

			file, ok := dir.Directories["aa"].Files[tC.file]
			if !ok {
				t.Fatalf(`no file "%s"`, tC.file)
			}

			err = file.Perform(rm, ctx)
			if err != nil {
				t.Fatal(err)
			}

			if file.OptionsContain("output") != tC.output {
				t.Fatal("output flag does not match")
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
