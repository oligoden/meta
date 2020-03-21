package entity_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/oligoden/meta/entity/state"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
)

func TestExecPerforming(t *testing.T) {
	testCases := []struct {
		desc    string
		cle     string
		prps    string
		content string
	}{
		{
			desc:    "normal exec",
			cle:     "cp",
			prps:    `"cmd":["cp", filepath.Join("a","aa.ext"), filepath.Join("a","ab.ext")]`,
			content: "a",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			str := `{
				"directories": {
					"a": {
						"directories": {
							"aa": {
								"files": {
									"aaa.ext": {}
								}
							}
						},
						"files": {
							"aa.ext": {}
						},
						"execs": {
							"%s": {%s}
						}
					}
				}
			}`
			str = fmt.Sprintf(str, tC.cle, tC.prps)

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
			rm.Evaluate()

			ctx := context.Background()
			ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/meta")
			ctx = context.WithValue(ctx, entity.ContextKey("destination"), "testing")
			ctx = context.WithValue(ctx, entity.ContextKey("watching"), false)
			ctx = context.WithValue(ctx, entity.ContextKey("force"), false)
			err = dir.Files["aa.ext"].Perform(ctx)
			if err != nil {
				t.Fatal(err)
			}
			err = dir.Execs["cp"].Perform(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := os.Stat("testing/a/ab.ext"); err != nil {
				t.Error(err)
			}

			rm.Finish()
			dir.Files["aa.ext"].State(state.Updated)
			rm.Propagate()

			exp := refmap.StatusText[state.Updated]
			got := refmap.StatusText[dir.Execs["cp"].State()]
			if got != exp {
				t.Errorf(`expected "%s", got "%s"`, exp, got)
			}
		})

		os.RemoveAll("testing/a/aa.ext")
		os.RemoveAll("testing/a/ab.ext")
	}
}
