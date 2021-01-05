package entity

import (
	"fmt"
)

type ProjectBranch struct {
	Project     string
	Testing     bool
	Environment string
	*Branch
}

func (pb *ProjectBranch) Build(e interface{}) (interface{}, error) {
	ent := e

	pb.Branch = &Branch{}
	ent, err := pb.Branch.Build(e)
	if err != nil {
		return nil, fmt.Errorf("building default branch, %w", err)
	}

	for {
		switch v := ent.(type) {
		case nil:
			return nil, fmt.Errorf("encountered nil")
		case *Project:
			pb.Project = v.Name
			pb.Testing = v.Testing
			pb.Environment = v.Environment
			return pb, nil
		default:
			return nil, fmt.Errorf("encountered unknown, %+v", v)
		}
	}
}
