package entity

import (
	"fmt"
	"regexp"
)

type controls struct {
	Behaviour *behaviour `json:"behaviour"`
	Mappings  []*Mapping `json:"mappings"`
}

type Mapping struct {
	Start      Regexp `json:"start"`
	End        Regexp `json:"end"`
	StartSet   string
	EndSet     string
	Recurrence int `json:"recurrence"`
}

type Regexp struct {
	regexp.Regexp
}

func (r *Regexp) UnmarshalText(text []byte) error {
	search := string(text)
	search = fmt.Sprintf("^%s$", search)
	re := regexp.MustCompile(`\.`)
	search = re.ReplaceAllString(search, `\.`)
	re = regexp.MustCompile(`\*`)
	search = re.ReplaceAllString(search, `.*`)

	rr, err := Compile(search)
	if err != nil {
		return err
	}
	*r = *rr
	return nil
}

func Compile(expr string) (*Regexp, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &Regexp{*re}, nil
}

const (
	NormalBehaviour = ""
	CopyBehaviour   = "copy"
	OutputBehaviour = "output"
)

type behaviour struct {
	Options    string  `json:"options"`
	Filters    filters `json:"filters"`
	Recurrence int     `json:"recurrence"`
}

type filters map[string]map[string]string

func (e Basic) ControlMappings() []*Mapping {
	return e.Controls.Mappings
}

func (b Basic) ContainsFilter(filter string) bool {
	if _, has := b.Controls.Behaviour.Filters[filter]; has {
		return true
	}
	return false
}

func (e Basic) Options() string {
	return e.Controls.Behaviour.Options
}

func (e Basic) Filters() filters {
	return e.Controls.Behaviour.Filters
}
