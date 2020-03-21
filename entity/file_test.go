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
	testCases := []struct {
		desc    string
		file    string
		prps    string
		dirCopy bool
		content string
	}{
		{
			desc:    "normal template execution",
			file:    "aaa.ext",
			content: "abc aaa.ext",
		},
		{
			desc:    "test source property",
			file:    "aaa.ext",
			prps:    `"source":"./aaz.ext"`,
			content: "def",
		},
		{
			desc:    "test source in sub directory",
			file:    "aaa.ext",
			prps:    `"source":"./sub/aax.ext"`,
			content: "ijk",
		},
		{
			desc:    "test removal of .tmpl",
			file:    "aaa.ext.tmpl",
			content: "ghi",
		},
		{
			desc:    "test copy only set on file",
			file:    "aaa.ext",
			prps:    `"copy-only":true`,
			content: "abc {{.Filename}}",
		},
		{
			desc:    "test copy only set on directory",
			file:    "aaa.ext",
			dirCopy: true,
			content: "abc {{.Filename}}",
		},
	}

	for _, tC := range testCases {
		dstFilename := strings.TrimSuffix("a/aa/"+tC.file, ".tmpl")

		t.Run(tC.desc, func(t *testing.T) {
			str := `{
				"directories": {
					"a": {
						"directories": {
							"aa": {
								"copy-only": %t,
								"files": {
									"%s": {%s}
								}
							}
						}
					}
				}
			}`
			str = fmt.Sprintf(str, tC.dirCopy, tC.file, tC.prps)

			m := &entity.Basic{}
			err := json.Unmarshal([]byte(str), &m)
			if err != nil {
				t.Error("error unmarshalling,", err)
			}

			dir := m.Directories["a"]
			dir.DestinationPath = "a"
			dir.SourcePath = "a"
			dir.Name = "a"
			m.Directories["a"].ParentID = "project:name"
			m.Directories["a"].Parent = m

			rm := refmap.Start()
			rm.AddRef("project:name", m)

			dir.Process(entity.BuildBranch, rm)
			file, ok := dir.Directories["aa"].Files[tC.file]
			if !ok {
				t.Fatalf(`no file "%s"`, tC.file)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/meta")
			ctx = context.WithValue(ctx, entity.ContextKey("destination"), "testing")
			ctx = context.WithValue(ctx, entity.ContextKey("watching"), false)
			ctx = context.WithValue(ctx, entity.ContextKey("force"), false)
			err = file.Perform(ctx)
			if err != nil {
				t.Fatal(err)
			}

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
		})

		os.RemoveAll("testing/" + dstFilename)
	}
}

func TestFileForcing(t *testing.T) {
	str := `{
		"directories": {
			"a": {
				"directories": {
					"aa": {
						"files": {
							"aab.ext": {}
						}
					}
				}
			}
		}
	}`
	str = fmt.Sprintf(str)

	m := &entity.Basic{}
	err := json.Unmarshal([]byte(str), &m)
	if err != nil {
		t.Error("error unmarshalling,", err)
	}

	dir := m.Directories["a"]
	dir.DestinationPath = "a"
	dir.SourcePath = "a"
	dir.Name = "a"
	m.Directories["a"].ParentID = "project:name"
	m.Directories["a"].Parent = m

	rm := refmap.Start()
	rm.AddRef("project:name", m)
	err = dir.Process(entity.BuildBranch, rm)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/meta")
	ctx = context.WithValue(ctx, entity.ContextKey("destination"), "testing")
	ctx = context.WithValue(ctx, entity.ContextKey("watching"), false)
	ctx = context.WithValue(ctx, entity.ContextKey("force"), true)

	f1, _ := os.Create("testing/a/aa/aab.ext")
	f1.WriteString("abc")
	f1.Close()

	f1, _ = os.Create("testing/meta/a/aa/aab.ext")
	f1.WriteString("def")
	f1.Close()

	err = dir.Directories["aa"].Files["aab.ext"].Perform(ctx)
	if err != nil {
		t.Fatal(err)
	}

	content, err := ioutil.ReadFile("testing/a/aa/aab.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp := "def"
	got := string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	os.RemoveAll("testing/meta/a/aa/aab.ext")
	os.RemoveAll("testing/a/aa/aab.ext")
}

func TestFileNotForcing(t *testing.T) {
	str := `{
		"directories": {
			"a": {
				"directories": {
					"aa": {
						"files": {
							"aab.ext": {}
						}
					}
				}
			}
		}
	}`
	str = fmt.Sprintf(str)

	m := &entity.Basic{}
	err := json.Unmarshal([]byte(str), &m)
	if err != nil {
		t.Error("error unmarshalling,", err)
	}

	dir := m.Directories["a"]
	dir.DestinationPath = "a"
	dir.SourcePath = "a"
	dir.Name = "a"
	m.Directories["a"].ParentID = "project:name"
	m.Directories["a"].Parent = m

	rm := refmap.Start()
	rm.AddRef("project:name", m)
	err = dir.Process(entity.BuildBranch, rm)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/meta")
	ctx = context.WithValue(ctx, entity.ContextKey("destination"), "testing")
	ctx = context.WithValue(ctx, entity.ContextKey("watching"), false)
	ctx = context.WithValue(ctx, entity.ContextKey("force"), false)

	f1, _ := os.Create("testing/a/aa/aab.ext")
	f1.WriteString("abc")
	f1.Close()

	f1, _ = os.Create("testing/meta/a/aa/aab.ext")
	f1.WriteString("def")
	f1.Close()

	err = dir.Directories["aa"].Files["aab.ext"].Perform(ctx)
	if err != nil {
		t.Fatal(err)
	}

	content, err := ioutil.ReadFile("testing/a/aa/aab.ext")
	if err != nil {
		t.Fatal(err)
	}
	exp := "abc"
	got := string(content)
	if exp != got {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	os.RemoveAll("testing/meta/a/aa/aab.ext")
	os.RemoveAll("testing/a/aa/aab.ext")
}
