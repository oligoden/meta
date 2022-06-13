package entity

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
)

type Directory struct {
	OrigOverride string `json:"orig"`
	DestOverride string `json:"dest"`
	// Import      *configImport `json:"import"`
	Basic
}

// type configImport struct {
// }

func (d Directory) Identifier() string {
	return "dir:" + d.SrcDerived + ":" + d.Name
}

func (d Directory) ContainsFilter(filter string) bool {
	if _, has := d.Flts[filter]; has {
		return true
	}
	return false
}

func (d Directory) Derived() (string, string) {
	return d.SrcDerived, d.DstDerived
}

func (d Directory) Output() string {
	return ""
}

func (d Directory) BehaviourOptionsContain(o string) bool {
	return strings.Contains(d.Opts, o)
}

func (e *Directory) Process(bb BranchBuilder, rm refmap.Mutator, ctx context.Context) error {
	SrcDerived, DstDerived := e.Parent.Derived()

	e.SrcDerived = filepath.Join(SrcDerived, e.Name)
	e.DstDerived = filepath.Join(DstDerived, e.Name)

	// verboseValue := ctx.Value(ContextKey("verbose")).(int)
	e.SrcDerived = path(e.SrcDerived, e.OrigOverride)
	e.DstDerived = path(e.DstDerived, e.DestOverride)

	e.This = e

	if e.Vars == nil {
		e.Vars = map[string]string{}
	}
	for k, v := range e.Parent.Variables() {
		if _, ok := e.Vars[k]; !ok {
			e.Vars[k] = v
		}
	}

	hash := ""
	nodes := rm.Nodes("", e.Identifier())
	if len(nodes) > 0 {
		hash = nodes[0].Hash()
	}
	e.Detect = state.New(hash)

	err := e.ProcessState()
	if err != nil {
		return err
	}

	err = e.Basic.Process(bb, rm, ctx)
	if err != nil {
		return err
	}

	return nil
}

func path(path, modify string) string {
	if strings.HasPrefix(modify, "../") {
		return filepath.Join(filepath.Dir(path), modify[3:])
	}
	if strings.HasPrefix(modify, "/") {
		return strings.TrimPrefix(modify, "/")
	}
	return filepath.Join(path, modify)
}

func (e *Directory) ProcessState() error {
	return e.Basic.ProcessState(e.OrigOverride + e.DestOverride)
}
