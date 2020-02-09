package project

import (
	"fmt"

	"github.com/oligoden/meta/entity"
)

type Branch struct {
	Project string
	entity.Branch
}

func BuildBranch(m entity.BranchSetter) (entity.UpStepper, error) {
	ent, err := entity.BuildBranch(m)
	if err != nil {
		return nil, err
	}

	b := Branch{
		Branch: m.SetBranch().(entity.Branch),
	}

	for {
		switch v := ent.(type) {
		case nil:
			m.SetBranch(b)
			return nil, fmt.Errorf("encountered nil UpStepper")
		case *Project:
			b.Project = v.Name
			m.SetBranch(b)
			return ent, nil
		default:
			return nil, fmt.Errorf("encountered unknown UpStepper, %+v", v)
		}
	}
}
