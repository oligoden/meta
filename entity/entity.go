package entity

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/oligoden/meta/entity/state"
)

type ContextKey string

type Basic struct {
	Name        string                `json:"name"`
	Directories map[string]*Directory `json:"directories"`
	ParentID    string                `json:"-"`
	Parent      UpStepper             `json:"-"`
	state.Detect
}

func (*Basic) Perform(context.Context) error {
	return nil
}

type UpStepper interface {
}

type Branch struct {
	Directories []string
	Filename    string
	TemplateMethods
}

type DataBranch interface {
}

type BranchSetter interface {
	SetBranch(...DataBranch) DataBranch
}

func BuildBranch(m BranchSetter) (UpStepper, error) {
	ent, ok := m.(UpStepper)
	if !ok {
		return nil, fmt.Errorf("not a UpStepper interface")
	}

	b := Branch{}
	b.Directories = []string{}

	for {
		switch v := ent.(type) {
		case nil:
			m.SetBranch(b)
			return nil, fmt.Errorf("encountered nil UpStepper")
		case *Directory:
			b.Directories = append(b.Directories, v.Name)
			ent = v.Parent
		case *file:
			b.Filename = v.Name
			ent = v.Parent
		default:
			m.SetBranch(b)
			return ent, nil
		}
	}
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
