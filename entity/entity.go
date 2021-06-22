package entity

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/inflection"
	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
)

type ContextKey string

type ConfigReader interface {
	Identifier() string
	Derived() (string, string)
	OptionsContain(string) bool
	ContainsFilter(string) bool
}

type Basic struct {
	Name        string                `json:"name"`
	SrcDerived  string                `json:"-"`
	DstDerived  string                `json:"-"`
	Directories map[string]*Directory `json:"directories"`
	Files       map[string]*file      `json:"files"`
	Execs       map[string]*cle       `json:"execs"`
	// Import      string                `json:"import"`
	Controls        controls     `json:"controls"`
	This            ConfigReader `json:"-"`
	Parent          ConfigReader `json:"-"`
	PosibleMappings []Mapping    `json:"-"`
	Settings        string       `json:"settings"`
	state.Detect
}

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

func (b Basic) Identifier() string {
	return b.Name
}

func (b Basic) ContainsFilter(filter string) bool {
	if _, has := b.Controls.Behaviour.Filters[filter]; has {
		return true
	}
	return false
}

func (b Basic) OptionsContain(o string) bool {
	if b.Controls.Behaviour == nil {
		return false
	}

	return strings.Contains(b.Controls.Behaviour.Options, o)
}

func (Basic) Perform(refmap.Grapher, context.Context) error {
	return nil
}

func (b Basic) Output() string {
	return ""
}

func (b Basic) Derived() (string, string) {
	return b.SrcDerived, b.DstDerived
}

func (n *Basic) Process(bb BranchBuilder, rm refmap.Mutator, ctx context.Context) error {
	if n.Controls.Behaviour == nil {
		n.Controls.Behaviour = &behaviour{}
	}
	return n.processFiles(bb, rm, ctx)
}

func (n *Basic) processFiles(bb BranchBuilder, rm refmap.Mutator, ctx context.Context) error {
	for name, file := range n.Files {
		file.Name = name
		file.Parent = n.This

		if file.Controls.Behaviour == nil {
			file.Controls.Behaviour = &behaviour{}

			if n.Controls.Behaviour != nil {
				file.Controls.Behaviour.Options = n.Controls.Behaviour.Options
				if file.Controls.Behaviour.Filters == nil {
					file.Controls.Behaviour.Filters = make(map[string]map[string]string)
				}
				for k, f := range n.Controls.Behaviour.Filters {
					file.Controls.Behaviour.Filters[k] = f
				}
			}
		}

		err := file.calculateHash()
		if err != nil {
			return err
		}

		_, err = bb.Build(file)
		if err != nil {
			return fmt.Errorf("building branch, %w", err)
		}
		file.Branch = bb

		if file.Source == "" {
			file.Source = filepath.Join(n.SrcDerived, name)
		} else if strings.HasPrefix(file.Source, "./") {
			file.Source = filepath.Join(n.SrcDerived, file.Source)
		}

		for _, m := range n.Controls.Mappings {
			matchStart := m.Start.MatchString(file.Identifier())
			matchEnd := m.End.MatchString(file.Identifier())
			if matchStart && matchEnd {
				return fmt.Errorf("directory matches start and end reference")
			}
			if matchStart {
				n.PosibleMappings = append(n.PosibleMappings, Mapping{
					StartSet:   file.Identifier(),
					End:        m.End,
					Recurrence: m.Recurrence,
				})
			}
			if matchEnd {
				n.PosibleMappings = append(n.PosibleMappings, Mapping{
					Start:      m.Start,
					EndSet:     file.Identifier(),
					Recurrence: m.Recurrence,
				})
			}
		}

		rm.AddRef("file:"+file.Source, file)
		err = rm.MapRef(file.Parent.Identifier(), "file:"+file.Source)
		if err != nil {
			return err
		}
	}

	return nil
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
		case *file:
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

//RandString creates a random string (https://stackoverflow.com/a/31832326)
func RandString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
