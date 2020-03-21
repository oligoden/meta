package entity_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
)

func TestDirectoryProcessing(t *testing.T) {
	testCases := []struct {
		desc              string
		dirSwitch         []string
		modDirSource      string
		modDirDestination string
		name              string
		fsp               string
		fdp               string
		filenames         []string
	}{
		{
			desc:      "test dir aa",
			dirSwitch: []string{"a", "aa"},
			name:      "aa",
			fsp:       "a/aa",
			fdp:       "a/aa",
			filenames: []string{"aaa.ext"},
		},
		{
			desc:      "test dir aaa",
			dirSwitch: []string{"a", "aa", "aaa"},
			name:      "aaa",
			fsp:       "a/aa/aaa",
			fdp:       "a/aa/aaa",
		},
		{
			desc:         "source stay in current",
			dirSwitch:    []string{"a", "aa"},
			modDirSource: ".",
			name:         "aa",
			fsp:          "a",
			fdp:          "a/aa",
		},
		{
			desc:         "source stay in current and add directory",
			dirSwitch:    []string{"a", "aa"},
			modDirSource: "./other",
			name:         "aa",
			fsp:          "a/other",
			fdp:          "a/aa",
		},
		{
			desc:         "source stay in current and add directory and test file",
			dirSwitch:    []string{"a", "aa"},
			modDirSource: "./other",
			name:         "aa",
			fsp:          "a/other",
			fdp:          "a/aa",
			filenames:    []string{"aaa.ext"},
		},
		{
			desc:         "source go to root",
			dirSwitch:    []string{"a", "aa"},
			modDirSource: "/",
			name:         "aa",
			fsp:          "",
			fdp:          "a/aa",
		},
		{
			desc:         "source go to root and add directory",
			dirSwitch:    []string{"a", "aa"},
			modDirSource: "/other",
			name:         "aa",
			fsp:          "other",
			fdp:          "a/aa",
		},
	}

	str := `{
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

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			m := entity.Basic{}
			json.Unmarshal([]byte(str), &m)
			dir := m.Directories["a"]
			var dirParent *entity.Directory

			m.Directories["a"].DestinationPath = "a"
			m.Directories["a"].SourcePath = "a"
			m.Directories["a"].Name = "a"
			m.Directories["a"].ParentID = "project:name"
			m.Directories["a"].Parent = m

			// switching to the directory you want to test with tC.dirs
			for _, dn := range tC.dirSwitch[1:] {
				dirTemp, found := dir.Directories[dn]
				if !found {
					t.Errorf("no directory '%s'", dn)
					t.FailNow()
				}
				dirParent = dir
				dir = dirTemp
			}

			rm := refMapStub{
				nodes: map[string]refmap.Actioner{},
				maps:  map[string]map[string]bool{},
			}

			dir.Source = tC.modDirSource
			dir.Destination = tC.modDirDestination
			err := m.Directories["a"].Process(entity.BuildBranch, rm)
			if err != nil {
				t.Error(err)
			}

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

			exp := "dir:" + dirParent.SourcePath + ":" + dirParent.Name
			got := dir.ParentID
			if got != exp {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}

			exp = tC.name
			got = dir.Name
			if got != exp {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}

			exp = tC.fsp
			got = dir.SourcePath
			if got != exp {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}

			exp = tC.fdp
			got = dir.DestinationPath
			if got != exp {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}

			exp1 := "dir:" + dir.SourcePath + ":" + dir.Name
			exp2 := "dir:" + dirParent.SourcePath + ":" + dirParent.Name
			if _, fnd := rm.nodes[exp1]; !fnd {
				t.Errorf(`expected to find "%s"`, exp1)
			}
			got1, fnd := rm.maps[exp2]
			if !fnd {
				t.Errorf(`expected to find "%s"`, exp2)
			}
			got2, fnd := got1[exp1]
			if !fnd {
				t.Errorf(`expected to find "%s"`, exp1)
			}
			if !got2 {
				t.Errorf(`expected true`)
			}

			if tC.filenames == nil {
				return
			}
			for _, fn := range tC.filenames {
				if _, fnd := dir.Files[fn]; !fnd {
					t.Fatal("expected to find", fn)
				}

				exp = fn
				got = dir.Files[fn].Name
				if got != exp {
					t.Errorf("expected '%s', got '%s'", exp, got)
				}

				exp = filepath.Join(dir.SourcePath, fn)
				got = dir.Files[fn].Identifier()
				if got != exp {
					t.Errorf("expected '%s', got '%s'", exp, got)
				}

				if d, ok := dir.Files[fn].Parent.(*entity.Directory); !ok {
					t.Error("parent is not a directory")
				} else {
					exp := dir.Hash()
					got := d.Hash()
					if got != exp {
						t.Errorf(`parent does not match, expected %+v, got %+v`, dirParent, d)
					}
				}

				exp = "dir:" + dir.SourcePath + ":" + dir.Name
				got = dir.Files[fn].ParentID
				if got != exp {
					t.Errorf(`expected "%s", got "%s"`, exp, got)
				}

				if dir.Files[fn].Hash() == "" {
					t.Errorf("expected hash, got empty string")
				}

				exp = tC.fsp + "/" + fn
				got = dir.Files[fn].Source
				if got != exp {
					t.Errorf("expected '%s', got '%s'", exp, got)
				}

				exp1 = "file:" + filepath.Join(dir.SourcePath, fn)
				exp2 = "dir:" + dir.SourcePath + ":" + dir.Name
				if _, fnd := rm.nodes[exp1]; !fnd {
					t.Errorf(`expected to find "%s"`, exp1)
				}
				got1, fnd := rm.maps[exp2]
				if !fnd {
					t.Errorf(`expected to find "%s"`, exp2)
				}
				got2, fnd := got1[exp1]
				if !fnd {
					t.Errorf(`expected to find "%s"`, exp2)
				}
				if !got2 {
					t.Errorf(`expected true`)
				}
			}
		})
	}
}

func TestDirectoryProcessingSame(t *testing.T) {
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

	rm := refMapStub{
		nodes: map[string]refmap.Actioner{},
		maps:  map[string]map[string]bool{},
	}
	m := entity.Basic{}
	json.Unmarshal([]byte(str), &m)
	m.Directories["a"].DestinationPath = "a"
	m.Directories["a"].SourcePath = "a"
	m.Directories["a"].Name = "a"
	m.Directories["a"].ParentID = "project:name"
	m.Directories["a"].Parent = m
	m.Directories["a"].Process(entity.BuildBranch, rm)
	hash := m.Directories["a"].Hash()

	m.Directories["a"].Files["aa.ext"].Copy = true
	m.Directories["a"].Directories["aa"].Copy = true
	m.Directories["a"].Process(entity.BuildBranch, rm)

	if m.Directories["a"].Hash() != hash {
		t.Error("expected hash to stay the same")
	}
}

func TestDirectoryProcessingUpdate(t *testing.T) {
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

	rm := refMapStub{
		nodes: map[string]refmap.Actioner{},
		maps:  map[string]map[string]bool{},
	}
	m := entity.Basic{}
	json.Unmarshal([]byte(str), &m)
	m.Directories["a"].DestinationPath = "a"
	m.Directories["a"].SourcePath = "a"
	m.Directories["a"].Name = "a"
	m.Directories["a"].ParentID = "project:name"
	m.Directories["a"].Parent = m
	m.Directories["a"].Process(entity.BuildBranch, rm)
	hash := m.Directories["a"].Hash()

	m.Directories["a"].Copy = true
	m.Directories["a"].Process(entity.BuildBranch, rm)

	if m.Directories["a"].Hash() == hash {
		t.Error("expected hash to change")
	}
}

func TestDirectoryProcessingLoadUpdate(t *testing.T) {
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

	rm := refMapStub{
		nodes: map[string]refmap.Actioner{},
		maps:  map[string]map[string]bool{},
	}
	m := entity.Basic{}
	json.Unmarshal([]byte(str), &m)
	m.Directories["a"].DestinationPath = "a"
	m.Directories["a"].SourcePath = "a"
	m.Directories["a"].Name = "a"
	m.Directories["a"].ParentID = "project:name"
	m.Directories["a"].Parent = m

	m.Directories["a"].Process(entity.BuildBranch, rm)
	hash := m.Directories["a"].Hash()

	str = `{
		"directories": {
			"a": {
				"copy-only": true,
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
	fmt.Printf("dir a, %+v\n", m.Directories["a"])
	json.Unmarshal([]byte(str), &m)
	fmt.Printf("dir a, %+v\n", m.Directories["a"])
	m.Directories["a"].Process(entity.BuildBranch, rm)

	if m.Directories["a"].Hash() == hash {
		t.Error("expected hash to change")
	}
}

type refMapStub struct {
	nodes map[string]refmap.Actioner
	maps  map[string]map[string]bool
}

func (rm refMapStub) AddRef(d string, f refmap.Actioner) {
	rm.nodes[d] = f
}

func (rm refMapStub) MapRef(a, b string, o ...uint) {
	if rm.maps[a] == nil {
		rm.maps[a] = map[string]bool{
			b: true,
		}
		return
	}
	rm.maps[a][b] = true
}
