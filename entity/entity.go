package entity

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/inflection"
	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
)

type ContextKey string

type Basic struct {
	Name        string                `json:"name"`
	Directories map[string]*Directory `json:"directories"`
	Execs       map[string]*cle       `json:"execs"`
	// Import      string                `json:"import"`
	Controls        controls   `json:"controls"`
	Parent          Identifier `json:"-"`
	PosibleMappings []Mapping  `json:"-"`
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
	rr, err := Compile(string(text))
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
)

type behaviour struct {
	Action     string  `json:"action"`
	Output     bool    `json:"output"`
	Filters    filters `json:"filters"`
	Recurrence int     `json:"recurrence"`
}

type filters map[string]map[string]string

func (fs filters) Has(f string) bool {
	if _, has := fs[f]; has {
		return true
	}
	return false
}

func (b Basic) Identifier() string {
	return b.Name
}

func (b Basic) Output() string {
	return ""
}

func (Basic) Perform(refmap.Grapher, context.Context) error {
	return nil
}

type Identifier interface {
	Identifier() string
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

	for {
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
