package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/oligoden/meta/refmap"
)

type Project struct {
	Basic
	Testing      bool       `json:"testing"`
	Environment  string     `json:"environment"`
	Repository   Repository `json:"repo"`
	WorkLocation string     `json:"work-location"`
	DestLocation string     `json:"dest-location"`
}

type Repository struct {
}

func Load(f io.Reader) (*Project, error) {
	p := &Project{}

	dec := json.NewDecoder(f)
	err := dec.Decode(p)
	if err != nil {
		return p, err
	}
	return p, nil
}

func (p *Project) Load(f io.Reader) error {
	dec := json.NewDecoder(f)
	err := dec.Decode(p)
	if err != nil {
		return err
	}
	return nil
}

func (p Project) Identifier() string {
	return "prj:" + p.Name
}

func (p Project) Derived() (string, string) {
	return p.SrcDerived, p.DstDerived
}

func (p *Project) Process(bb BranchBuilder, rm refmap.Mutator, ctx context.Context) error {

	err := p.calculateHash()
	if err != nil {
		return err
	}

	rm.AddRef(p.Identifier(), p)

	for name, dir := range p.Directories {
		dir.Name = name
		dir.Parent = p
		dir.SrcDerived = name
		dir.DstDerived = name

		if dir.Controls.Behaviour == nil {
			dir.Controls.Behaviour = &behaviour{}
			if p.Controls.Behaviour != nil {
				dir.Controls.Behaviour.Options = p.Controls.Behaviour.Options
				if dir.Controls.Behaviour.Filters == nil {
					dir.Controls.Behaviour.Filters = make(map[string]map[string]string)
				}
				for k, f := range p.Controls.Behaviour.Filters {
					dir.Controls.Behaviour.Filters[k] = f
				}
			}
		}

		for _, m := range p.Controls.Mappings {
			matchStart := m.Start.MatchString(dir.Identifier())
			matchEnd := m.End.MatchString(dir.Identifier())
			if matchStart && matchEnd {
				return fmt.Errorf("directory matches start and end reference")
			}
			if matchStart {
				p.PosibleMappings = append(p.PosibleMappings, Mapping{
					StartSet:   dir.Identifier(),
					End:        m.End,
					Recurrence: m.Recurrence,
				})
			}
			if matchEnd {
				p.PosibleMappings = append(p.PosibleMappings, Mapping{
					Start:      m.Start,
					EndSet:     dir.Identifier(),
					Recurrence: m.Recurrence,
				})
			}
		}
	}

	p.This = p
	err = p.Basic.Process(bb, rm, ctx)
	if err != nil {
		return err
	}

	for name, e := range p.Execs {
		e.Name = name
		e.Parent = p

		err := e.calculateHash()
		if err != nil {
			return err
		}

		for _, m := range p.Controls.Mappings {
			matchStart := m.Start.MatchString(e.Identifier())
			matchEnd := m.End.MatchString(e.Identifier())
			if matchStart && matchEnd {
				return fmt.Errorf("directory matches start and end reference")
			}
			if matchStart {
				p.PosibleMappings = append(p.PosibleMappings, Mapping{
					StartSet:   e.Identifier(),
					End:        m.End,
					Recurrence: m.Recurrence,
				})
			}
			if matchEnd {
				p.PosibleMappings = append(p.PosibleMappings, Mapping{
					Start:      m.Start,
					EndSet:     e.Identifier(),
					Recurrence: m.Recurrence,
				})
			}
		}

		rm.AddRef(e.Identifier(), e)
		err = rm.MapRef(p.Identifier(), e.Identifier())
		if err != nil {
			return err
		}
	}

	// look for directories matching posible mappings
	for _, dir := range p.Directories {
		for _, m := range p.PosibleMappings {
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
	for _, file := range p.Files {
		for _, m := range p.PosibleMappings {
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
	for _, e := range p.Execs {
		for _, m := range p.PosibleMappings {
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

	for _, dir := range p.Directories {
		for _, m := range p.Controls.Mappings {
			// 	if m.Recurrence > 0 {
			// 		m.Recurrence--
			// 		dir.Controls.Mappings = append(dir.Controls.Mappings, m)
			// 	}
			// 	if m.Recurrence == -1 {
			// 		dir.Controls.Mappings = append(dir.Controls.Mappings, m)
			// 	}
			dir.Controls.Mappings = append(dir.Controls.Mappings, m)
		}

		// for _, m := range p.PosibleMappings {
		// 	if m.Recurrence > 0 {
		// 		m.Recurrence--
		// 		dir.PosibleMappings = append(dir.PosibleMappings, m)
		// 	}
		// 	if m.Recurrence == -1 {
		// 		dir.PosibleMappings = append(dir.PosibleMappings, m)
		// 	}
		// }

		dir.PosibleMappings = p.PosibleMappings
		err := dir.Process(bb, rm, ctx)
		if err != nil {
			return err
		}
		p.PosibleMappings = dir.PosibleMappings
	}

	return nil
}

func (p *Project) calculateHash() error {
	pTemp := *p
	pTemp.Directories = nil

	err := p.HashOf(pTemp)
	if err != nil {
		return err
	}
	return nil
}
