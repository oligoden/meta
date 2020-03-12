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
)

type ContextKey string

type Basic struct {
	Name        string                `json:"name"`
	RS          string                `json:"-"`
	Directories map[string]*Directory `json:"directories"`
	Execs       map[string]*cle       `json:"execs"`
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
