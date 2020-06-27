package entity

import (
	"fmt"
)

type ProjectBranch struct {
	Project     string
	Testing     bool
	Environment string
	Branch
}

func BuildProjectBranch(m BranchSetter) (UpStepper, error) {
	ent, err := BuildBranch(m)
	if err != nil {
		return nil, err
	}

	b := ProjectBranch{
		Branch: m.SetBranch().(Branch),
	}

	for {
		switch v := ent.(type) {
		case nil:
			m.SetBranch(b)
			return nil, fmt.Errorf("encountered nil UpStepper")
		case *Project:
			b.Project = v.Name
			b.Testing = v.Testing
			b.Environment = v.Environment
			m.SetBranch(b)
			return ent, nil
		default:
			return nil, fmt.Errorf("encountered unknown UpStepper, %+v", v)
		}
	}
}
