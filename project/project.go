package project

import (
	"context"
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

func (p *Project) Process(bb func(entity.BranchSetter) (entity.UpStepper, error), m refmap.Mutator) error {
	err := p.calculateHash()
	if err != nil {
		return err
	}

	for name, dir := range p.Directories {
		dir.Name = name
		dir.Parent = p
		dir.SourcePath = name
		dir.DestinationPath = name
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

func Build(ctx context.Context, m *refmap.Store) {
	for _, ref := range m.ChangedRefs() {
		for _, val := range ref.Files {
			val.Perform(ctx)
		}
	}
}
