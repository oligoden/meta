// +build linux darwin

package entity_test

import (
	"context"
	"os"
	"testing"

	"github.com/oligoden/meta/entity"
)

func TestExec(t *testing.T) {
	t.Skip()
	cfg := `{
		"directories": {
			"a": {
				"files": {
					"aa.ext": {}
				},
				"execs": {
					"cp": {
						"cmd": ["cp", "a/aa.ext", "a/ab.ext"]
					}
				}
			}
		}
	}`

	mc := dirTestConfigSetup(cfg, t)
	dir := mc.Directories["a"]
	ctx := context.Background()
	ctx = context.WithValue(ctx, entity.ContextKey("source"), "testing/work")
	ctx = context.WithValue(ctx, entity.ContextKey("destination"), "testing")
	ctx = context.WithValue(ctx, entity.ContextKey("watching"), false)
	ctx = context.WithValue(ctx, entity.ContextKey("force"), false)
	ctx = context.WithValue(ctx, entity.ContextKey("verbose"), 0)
	rm := dirTestProcess(mc, ctx, t)

	exec := dir.Execs["cp"]

	exp := "exec:cp"
	got := exec.Identifier()
	if got != exp {
		t.Errorf(`expected "%s", got "%s"`, exp, got)
	}

	err := dir.Files["aa.ext"].Perform(rm, ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = dir.Execs["cp"].Perform(rm, ctx)
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat("testing/a/ab.ext"); err != nil {
		t.Error(err)
	}

	os.RemoveAll("testing/a/aa.ext")
	os.RemoveAll("testing/a/ab.ext")
}
