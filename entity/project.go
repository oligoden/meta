package entity

import (
	"context"
	"encoding/json"
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

func (p *Project) Process(bb func(BranchSetter) (UpStepper, error), m refmap.Mutator, ctx context.Context) error {
	p.Edges = []Edge{}

	err := p.calculateHash()
	if err != nil {
		return err
	}
	m.AddRef("project", p)

	cleLinks := []string{}
	for name, e := range p.Execs {
		e.Name = name
		e.Parent = p
		e.ParentID = "project"
		e.Process()

		m.AddRef("exec:"+name, e)
		err = m.MapRef("project", "exec:"+name)
		if err != nil {
			return err
		}
		cleLinks = append(cleLinks, "exec:"+name)
	}

	for name, dir := range p.Directories {
		dir.Name = name
		dir.Parent = p
		dir.ParentID = "project"
		dir.SourcePath = name
		dir.DestinationPath = name
		dir.LinkTo = cleLinks
		dir.Edges = p.Edges
		err := dir.Process(bb, m, ctx)
		p.Edges = dir.Edges
		if err != nil {
			return err
		}
	}

	for _, e := range p.Edges {
		err = m.MapRef(e.Start, e.End)
		if err != nil {
			return err
		}
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
