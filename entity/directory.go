package entity

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/oligoden/meta/refmap"
	"gopkg.in/fsnotify.v1"
)

type Directory struct {
	SrcOverride string        `json:"src-ovr"`
	DstOverride string        `json:"dst-ovr"`
	Import      *configImport `json:"import"`

	// Settings can contain:
	// - "copy-only" to only copy file
	// - "parse-dir" to parse all templates in directory
	// - "comment-filter" to apply comment line filter
	// - "no-output" to skip file output

	LinkTo []string `json:"linkto"`
	Basic
}

type configImport struct {
}

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

func (d *Directory) calculateHash() error {
	dirTemp := *d
	dirTemp.Directories = nil
	dirTemp.Files = nil
	err := d.HashOf(dirTemp)
	if err != nil {
		return err
	}
	return nil
}

func (d *Directory) Process(bb BranchBuilder, rm refmap.Mutator, ctx context.Context) error {
	// verboseValue := ctx.Value(ContextKey("verbose")).(int)
	d.SrcDerived = path(d.SrcDerived, d.SrcOverride)
	d.DstDerived = path(d.DstDerived, d.DstOverride)

	if d.Import != nil {
		rootSrcDir := ctx.Value(ContextKey("source")).(string)
		metafile := filepath.Join(rootSrcDir, d.SrcDerived, "/meta.json")

		f, err := os.Open(metafile)
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer f.Close()

		p, err := Load(f)
		if err != nil {
			fmt.Println(err)
			return err
		}

		if metfileWatcher, ok := ctx.Value(ContextKey("watcher")).(*fsnotify.Watcher); ok {
			metfileWatcher.Add(metafile)
		}

		if d.Directories == nil {
			d.Directories = map[string]*Directory{}
		}

		for k, v := range p.Directories {
			d.Directories[k] = v
			updatePaths(d.Directories[k], d.SrcDerived)
		}
	}

	err := d.calculateHash()
	if err != nil {
		return err
	}

	rm.AddRef(d.Identifier(), d)
	err = rm.MapRef(d.Parent.Identifier(), d.Identifier())
	if err != nil {
		return fmt.Errorf("mapping reference, %w", err)
	}

	for name, dir := range d.Directories {
		dir.Name = name
		dir.Parent = d
		dir.SrcDerived = filepath.Join(d.SrcDerived, name)
		dir.DstDerived = filepath.Join(d.DstDerived, name)

		if dir.Controls.Behaviour == nil {
			dir.Controls.Behaviour = &behaviour{}
			if d.Controls.Behaviour != nil {
				dir.Controls.Behaviour.Options = d.Controls.Behaviour.Options
				if dir.Controls.Behaviour.Filters == nil {
					dir.Controls.Behaviour.Filters = make(map[string]map[string]string)
				}
				for k, f := range d.Controls.Behaviour.Filters {
					dir.Controls.Behaviour.Filters[k] = f
				}
			}
		}

		for _, m := range d.Controls.Mappings {
			matchStart := m.Start.MatchString(dir.Identifier())
			matchEnd := m.End.MatchString(dir.Identifier())
			if matchStart && matchEnd {
				return fmt.Errorf("directory matches start and end reference")
			}
			if matchStart {
				d.PosibleMappings = append(d.PosibleMappings, Mapping{
					StartSet:   dir.Identifier(),
					End:        m.End,
					Recurrence: m.Recurrence,
				})
			}
			if matchEnd {
				d.PosibleMappings = append(d.PosibleMappings, Mapping{
					Start:      m.Start,
					EndSet:     dir.Identifier(),
					Recurrence: m.Recurrence,
				})
			}
		}
	}

	d.This = d
	err = d.Basic.Process(bb, rm, ctx)
	if err != nil {
		return err
	}

	for name, e := range d.Execs {
		e.Name = name
		e.Parent = d

		err := e.calculateHash()
		if err != nil {
			return err
		}

		for _, m := range d.Controls.Mappings {
			matchStart := m.Start.MatchString(e.Identifier())
			matchEnd := m.End.MatchString(e.Identifier())
			if matchStart && matchEnd {
				return fmt.Errorf("directory matches start and end reference")
			}
			if matchStart {
				d.PosibleMappings = append(d.PosibleMappings, Mapping{
					StartSet:   e.Identifier(),
					End:        m.End,
					Recurrence: m.Recurrence,
				})
			}
			if matchEnd {
				d.PosibleMappings = append(d.PosibleMappings, Mapping{
					Start:      m.Start,
					EndSet:     e.Identifier(),
					Recurrence: m.Recurrence,
				})
			}
		}

		rm.AddRef("exec:"+name, e)
		err = rm.MapRef(d.Identifier(), "exec:"+name)
		if err != nil {
			return err
		}
	}

	// look for directories matching posible mappings
	for _, dir := range d.Directories {
		for _, m := range d.PosibleMappings {
			if m.StartSet == "" && dir.Identifier() != m.EndSet {
				match := m.Start.MatchString(dir.Identifier())
				if match {
					err := rm.MapRef(dir.Identifier(), m.EndSet)
					if err != nil {
						return err
					}
				}
			}
			if m.EndSet == "" && dir.Identifier() != m.StartSet {
				match := m.End.MatchString(dir.Identifier())
				if match {
					err := rm.MapRef(m.StartSet, dir.Identifier())
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// look for files matching posible mappings
	for _, file := range d.Files {
		for _, m := range d.PosibleMappings {
			if m.StartSet == "" && file.Identifier() != m.EndSet {
				match := m.Start.MatchString(file.Identifier())
				if match {
					err := rm.MapRef(file.Identifier(), m.EndSet)
					if err != nil {
						return err
					}
				}
			}
			if m.EndSet == "" && file.Identifier() != m.StartSet {
				match := m.End.MatchString(file.Identifier())
				if match {
					err := rm.MapRef(m.StartSet, file.Identifier())
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// look for execs matching posible mappings
	for _, e := range d.Execs {
		for _, m := range d.PosibleMappings {
			if m.StartSet == "" && e.Identifier() != m.EndSet {
				match := m.Start.MatchString(e.Identifier())
				if match {
					err := rm.MapRef(e.Identifier(), m.EndSet)
					if err != nil {
						return err
					}
				}
			}
			if m.EndSet == "" && e.Identifier() != m.StartSet {
				match := m.End.MatchString(e.Identifier())
				if match {
					err := rm.MapRef(m.StartSet, e.Identifier())
					if err != nil {
						return err
					}
				}
			}
		}
	}

	for _, dir := range d.Directories {
		for _, m := range d.Controls.Mappings {
			// if m.Recurrence > 0 {
			// 	m.Recurrence--
			// 	dir.Controls.Mappings = append(dir.Controls.Mappings, m)
			// }
			// if m.Recurrence == -1 {
			// 	dir.Controls.Mappings = append(dir.Controls.Mappings, m)
			// }
			dir.Controls.Mappings = append(dir.Controls.Mappings, m)
		}

		// for _, m := range d.PosibleMappings {
		// 	if m.Recurrence > 0 {
		// 		m.Recurrence--
		// 		dir.PosibleMappings = append(dir.PosibleMappings, m)
		// 	}
		// 	if m.Recurrence == -1 {
		// 		dir.PosibleMappings = append(dir.PosibleMappings, m)
		// 	}
		// }

		dir.PosibleMappings = d.PosibleMappings
		err := dir.Process(bb, rm, ctx)
		if err != nil {
			return err
		}
		d.PosibleMappings = dir.PosibleMappings
	}

	return nil
}

func path(path, modify string) string {
	if strings.HasPrefix(modify, ".") {
		return filepath.Join(filepath.Dir(path), modify)
	}
	if strings.HasPrefix(modify, "/") {
		return strings.TrimPrefix(modify, "/")
	}
	return filepath.Join(path, modify)
}

func updatePaths(d *Directory, p string) {
	for i, file := range d.Files {
		for j, t := range file.Templates {
			d.Files[i].Templates[j] = filepath.Join(p, t)
		}
	}

	for k := range d.Directories {
		updatePaths(d.Directories[k], p)
	}
}
