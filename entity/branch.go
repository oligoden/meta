package entity

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jinzhu/inflection"
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

type Branch struct {
	Directories []string
	Filename    string
	TemplateMethods
}

type BranchBuilder interface {
	Build(interface{}) (interface{}, error)
}

func (b *Branch) Build(e interface{}) (interface{}, error) {
	ent := e

	if b.Directories == nil {
		b.Directories = []string{}
	}

	for i := 0; i < 40; i++ {
		switch v := ent.(type) {
		case nil:
			return nil, fmt.Errorf("encountered nil")
		case *Directory:
			b.Directories = append(b.Directories, v.Name)
			ent = v.Parent
		case *File:
			b.Filename = v.Name
			ent = v.Parent
		default:
			return ent, nil
		}
	}

	fmt.Printf("branch depth exceeded, stuck at %+v\n", ent)
	return ent, fmt.Errorf("branch depth exceeded")
}

type TemplateMethods struct {
}

func (TemplateMethods) Clean(s string) string {
	reg, _ := regexp.Compile("[^a-zA-Z]+")
	return reg.ReplaceAllString(s, "")
}

func (TemplateMethods) Upper(s string) string {
	return strings.ToUpper(s)
}

func (TemplateMethods) CleanUpper(s string) string {
	reg, _ := regexp.Compile("[^a-zA-Z]+")
	clean := reg.ReplaceAllString(s, "")
	return strings.ToUpper(clean)
}

func (TemplateMethods) Title(s string) string {
	return strings.Title(s)
}

func (TemplateMethods) Plural(s string) string {
	return inflection.Plural(s)
}

func (TemplateMethods) Env(s string) string {
	return os.Getenv(s)
}
