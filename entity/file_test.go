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

func TestFileProcessing(t *testing.T) {
	testCases := []struct {
		desc   string
		id     string
		name   string
		source string
	}{
		{
			desc:   "normal file",
			id:     "aa.ext",
			name:   "aa.ext",
			source: "aa.ext",
		},
		{
			desc:   "file with other source",
			id:     "ab.ext",
			name:   "ab.ext",
			source: "other.ext",
		},
	}

	str := `{
		"directories": {
			"a": {
				"files": {
					"aa.ext": {},
					"ab.ext": {
						"source": "other.ext"
					}
				}
			}
		}
	}`

	m := entity.Basic{}
	err := json.Unmarshal([]byte(str), &m)
	if err != nil {
		t.Error("error unmarshalling")
	}

	dir := m.Directories["a"]
	dir.DestinationPath = "a"
	dir.SourcePath = "a"
	dir.Name = "a"
	dir.Parent = struct{}{}
	rm := refMapStub{}
	err = dir.Process(entity.BuildBranch, rm)
	if err != nil {
		t.Errorf("process error, %v", err)
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := dir.Files[tC.id].Hash()
			if len(got) == 0 {
				t.Errorf("expected hash, got empty string")
			}

			exp := tC.name
			got = dir.Files[tC.id].Name
			if got != exp {
				t.Errorf("expected '%s', got '%s'\n", exp, got)
			}

			if d, ok := dir.Files[tC.id].Parent.(*entity.Directory); !ok {
				t.Error("parent is not a directory")
			} else {
				exp := dir.Hash()
				got := d.Hash()
				if got != exp {
					t.Errorf(`parent does not match, expected %+v, got %+v`, dir, d)
				}
			}

			exp = "a/" + tC.name
			got = rm["a/"+tC.name].destination
			if got != exp {
				t.Errorf("expected '%s', got '%s'\n", exp, got)
			}
		})
	}
}

type refMapStub map[string]struct {
	destination string
	file        refmap.Actioner
}

func (rm refMapStub) Write(s, d string, f refmap.Actioner) {
	fmt.Println("write", s, d)
	rm[s] = struct {
		destination string
		file        refmap.Actioner
	}{
		destination: d,
	}
}

func TestFilePerforming(t *testing.T) {
	testPath := "testing/meta"

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
			prps:    `"source":"aaz.ext"`,
			content: "def",
		},
		{
			desc:    "test source in sub directory",
			file:    "aaa.ext",
			prps:    `"source":"sub/aax.ext"`,
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
			str := `{"directories": {"a": {"directories": {"aa": {"copy-only": %t, "files": {"%s": {%s}}}}}}}`
			str = fmt.Sprintf(str, tC.dirCopy, tC.file, tC.prps)

			m := entity.Basic{}
			err := json.Unmarshal([]byte(str), &m)
			if err != nil {
				t.Error("error unmarshalling,", err)
			}

			dir := m.Directories["a"]
			dir.DestinationPath = "a"
			dir.SourcePath = "a"
			dir.Name = "a"
			rm := refmap.Start(testPath)
			dir.Process(entity.BuildBranch, rm)
			file, ok := dir.Directories["aa"].Files[tC.file]
			if !ok {
				t.Fatalf("no file '%s'", tC.file)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, entity.ContextKey("source"), testPath)
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

func Test(t *testing.T) {
	testCases := []struct {
		desc  string
		force bool
		exp   string
	}{
		{
			desc:  "without force",
			force: false,
			exp:   "abc",
		},
		{
			desc:  "with force",
			force: true,
			exp:   "def",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			os.Mkdir("testing/meta", os.ModePerm)
			f1, _ := os.Create("testing/meta/a/aa.ext")
			f1.WriteString("abc")
			f1.Close()

			str := `{"directories": {"a": {"files": {"aa.ext": {}}}}}`

			m := entity.Basic{}
			err := json.Unmarshal([]byte(str), &m)
			if err != nil {
				t.Error("error unmarshalling,", err)
			}

			dir := m.Directories["a"]
			dir.DestinationPath = "a"
			dir.SourcePath = "a"
			dir.Name = "a"
			rm := refmap.Start("testing/meta")
			dir.Process(entity.BuildBranch, rm)
			file := dir.Files["aa.ext"]

			ctx := context.Background()
			ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/meta")
			ctx = context.WithValue(ctx, entity.ContextKey("destination"), "testing")
			ctx = context.WithValue(ctx, entity.ContextKey("watching"), true)
			ctx = context.WithValue(ctx, entity.ContextKey("force"), tC.force)

			err = file.Perform(ctx)
			if err != nil {
				t.Fatal(err)
			}

			f1, _ = os.Create("testing/meta/a/aa.ext")
			f1.WriteString("def")
			f1.Close()

			err = file.Perform(ctx)
			if err != nil {
				t.Fatal(err)
			}

			content, err := ioutil.ReadFile("testing/a/aa.ext")
			if err != nil {
				t.Fatal(err)
			}
			exp := tC.exp
			got := string(content)
			if exp != got {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}

			os.RemoveAll("testing/meta/a/aa.ext")
			os.RemoveAll("testing/a/aa.ext")
		})
	}
}
