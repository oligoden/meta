package project

import (
	"encoding/json"
	"io"

	"github.com/oligoden/meta/refmap"

	"github.com/oligoden/meta/entity"
)

type Project struct {
	entity.Basic
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

func (p *Project) Load(f io.Reader) (*Project, error) {
	dec := json.NewDecoder(f)
	err := dec.Decode(p)
	if err != nil {
		return p, err
	}
	return p, nil
}

func (p *Project) Process(bb func(entity.BranchSetter) (entity.UpStepper, error), m refmap.Mutator) error {
	err := p.calculateHash()
	if err != nil {
		return err
	}
	m.AddRef("project", p)

	cleLinks := []string{}
	for name, e := range p.Execs {
		e.Parent = p
		e.ParentID = "project"
		e.Process()

		m.AddRef("exec:"+name, e)
		m.MapRef("project", "exec:"+name)
		cleLinks = append(cleLinks, "exec:"+name)
	}

	for name, dir := range p.Directories {
		dir.Name = name
		dir.Parent = p
		dir.ParentID = "project"
		dir.SourcePath = name
		dir.DestinationPath = name
		dir.LinkTo = cleLinks
		err := dir.Process(bb, m)
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
