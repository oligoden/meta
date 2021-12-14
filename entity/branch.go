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
	Branch
}

func (pb *ProjectBranch) Build(e interface{}) (interface{}, error) {
	ent, err := (&pb.Branch).Build(e)
	if err != nil {
		return nil, fmt.Errorf("building default branch -> %w", err)
	}

	switch v := ent.(type) {
	case *Project:
		pb.Project = v.Name
		pb.Testing = v.Testing
		pb.Environment = v.Environment
		return ent, nil
	default:
		return ent, fmt.Errorf("encountered unknown, %+v", v)
	}
}

func (b *ProjectBranch) Clone() BranchBuilder {
	t := *b
	return &t
}

type Branch struct {
	Directories []string
	Filename    string
	Vars        map[string]string
	TemplateMethods
}

type BranchBuilder interface {
	Build(interface{}) (interface{}, error)
	Clone() BranchBuilder
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
			b.Vars = v.Vars
			ent = v.Parent
		case *File:
			b.Filename = v.Name
			b.Vars = v.Vars
			ent = v.Parent
		default:
			return ent, nil
		}
	}

	return ent, fmt.Errorf("branch depth exceeded")
}

func (b *Branch) Clone() BranchBuilder {
	t := *b
	return &t
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
