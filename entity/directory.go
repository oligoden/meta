package entity

import (
	"context"
	"log"
	"path/filepath"
	"strings"

	"github.com/oligoden/meta/refmap"
)

type Directory struct {
	SrcOverride string `json:"src-ovr"`
	DstOverride string `json:"dst-ovr"`
	// Import      *configImport `json:"import"`
	Basic
}

// type configImport struct {
// }

func (d Directory) Identifier() string {
	return "dir:" + d.SrcDerived + ":" + d.Name
}

func (d Directory) ContainsFilter(filter string) bool {
	if _, has := d.Controls.Behaviour.Filters[filter]; has {
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
	if d.Controls.Behaviour == nil {
		log.Println("behaviour not set for", d.Name)
		return false
	}

	return strings.Contains(d.Controls.Behaviour.Options, o)
}

func (d *Directory) HashOf() error {
	dirTemp := *d
	dirTemp.Directories = nil
	dirTemp.Files = nil
	err := d.Basic.HashOf()
	if err != nil {
		return err
	}
	return nil
}

func (e *Directory) Process(bb BranchBuilder, rm refmap.Mutator, ctx context.Context) error {
	SrcDerived, DstDerived := e.Parent.Derived()

	e.SrcDerived = filepath.Join(SrcDerived, e.Name)
	e.DstDerived = filepath.Join(DstDerived, e.Name)

	// verboseValue := ctx.Value(ContextKey("verbose")).(int)
	e.SrcDerived = path(e.SrcDerived, e.SrcOverride)
	e.DstDerived = path(e.DstDerived, e.DstOverride)

	e.This = e

	err := e.Basic.Process(bb, rm, ctx)
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

// func updatePaths(d *Directory, p string) {
// 	for i, file := range d.Files {
// 		for j, t := range file.Templates {
// 			d.Files[i].Templates[j] = filepath.Join(p, t)
// 		}
// 	}

// 	for k := range d.Directories {
// 		updatePaths(d.Directories[k], p)
// 	}
// }
