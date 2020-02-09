package entity_test

import (
	"encoding/json"
	"testing"

	"github.com/oligoden/meta/entity"
)

func TestDirectoryProcessing(t *testing.T) {
	testCases := []struct {
		desc           string
		name           string
		dirs           []string
		fsp            string
		fdp            string
		dirSource      string
		dirDestination string
	}{
		{
			desc: "test dir aa",
			name: "aa",
			dirs: []string{"a", "aa"},
			fsp:  "a/aa",
			fdp:  "a/aa",
		},
		{
			desc: "test dir aaa",
			name: "aaa",
			dirs: []string{"a", "aa", "aaa"},
			fsp:  "a/aa/aaa",
			fdp:  "a/aa/aaa",
		},
		{
			desc:      "source stay in current",
			name:      "aa",
			dirs:      []string{"a", "aa"},
			fsp:       "a",
			fdp:       "a/aa",
			dirSource: ".",
		},
		{
			desc:      "source stay in current and add directory",
			name:      "aa",
			dirs:      []string{"a", "aa"},
			fsp:       "a/other",
			fdp:       "a/aa",
			dirSource: "./other",
		},
		{
			desc:      "source go to root",
			name:      "aa",
			dirs:      []string{"a", "aa"},
			fsp:       "",
			fdp:       "a/aa",
			dirSource: "/",
		},
		{
			desc:      "source go to root and add directory",
			name:      "aa",
			dirs:      []string{"a", "aa"},
			fsp:       "other",
			fdp:       "a/aa",
			dirSource: "/other",
		},
	}

	str := `{
		"directories": {
			"a": {
				"directories": {
					"aa": {
						"directories": {
							"aaa": {}
						}
					}
				},
				"files": {
					"aa.ext": {

					}
				}
			}
		}
	}`

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m := entity.Basic{}
			json.Unmarshal([]byte(str), &m)
			dir := m.Directories["a"]
			var dirParent *entity.Directory

			m.Directories["a"].DestinationPath = "a"
			m.Directories["a"].SourcePath = "a"
			m.Directories["a"].Name = "a"

			// switching to the directory you want to test with tC.dirs
			for _, dn := range tC.dirs[1:] {
				dirTemp, found := dir.Directories[dn]
				if !found {
					t.Errorf("no directory '%s'", dn)
					t.FailNow()
				}
				dirParent = dir
				dir = dirTemp
			}

			dir.Source = tC.dirSource
			dir.Destination = tC.dirDestination
			m.Directories["a"].Process(entity.BuildBranch, nil)

			if dir.Hash() == "" {
				t.Errorf("expected hash, got empty string")
			}

			if d, ok := dir.Parent.(*entity.Directory); !ok {
				t.Error("parent is not a directory")
			} else {
				exp := dirParent.Hash()
				got := d.Hash()
				if got != exp {
					t.Errorf(`parent does not match, expected %+v, got %+v`, dirParent, d)
				}
			}

			exp := tC.name
			got := dir.Name
			if got != exp {
				t.Errorf("expected '%s', got '%s'", exp, got)
			}

			exp = tC.fsp
			got = dir.SourcePath
			if got != exp {
				t.Errorf("expected '%s', got '%s'", exp, got)
			}

			exp = tC.fdp
			got = dir.DestinationPath
			if got != exp {
				t.Errorf("expected '%s', got '%s'", exp, got)
			}
		})
	}
}

func TestDirectoryProcessingHashCalc(t *testing.T) {
	str := `{
		"directories": {
			"a": {
				"directories": {
					"aa": {
						
					}
				},
				"files": {
					"aa.ext": {

					}
				}
			}
		}
	}`

	m := entity.Basic{}
	json.Unmarshal([]byte(str), &m)
	m.Directories["a"].DestinationPath = "a"
	m.Directories["a"].SourcePath = "a"
	m.Directories["a"].Name = "a"
	m.Directories["a"].Process(entity.BuildBranch, nil)

	hash := m.Directories["a"].Hash()

	m.Directories["a"].Files["aa.ext"].Copy = true
	m.Directories["a"].Directories["aa"].Copy = true
	m.Directories["a"].Process(entity.BuildBranch, nil)

	if m.Directories["a"].Hash() != hash {
		t.Error("expected hash to stay the same")
	}
}
