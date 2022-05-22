package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
	"gopkg.in/fsnotify.v1"
)

type ConfigReader interface {
	Identifier() string
	Derived() (string, string)
	ControlMappings() []*Mapping
	AddPosibleMapping(Mapping)
	Options() string
	ContainsFilter(string) bool
	Filters() filters
	Variables() map[string]string
	refmap.Actioner
}

type Basic struct {
	Name            string                `json:"name"`
	SrcDerived      string                `json:"-"`
	DstDerived      string                `json:"-"`
	Vars            map[string]string     `json:"vars"`
	Directories     map[string]*Directory `json:"directories"`
	Files           map[string]*File      `json:"files"`
	Execs           map[string]*CLE       `json:"execs"`
	Import          bool                  `json:"import"`
	Controls        Controls              `json:"controls"`
	This            ConfigReader          `json:"-"`
	Parent          ConfigReader          `json:"-"`
	posibleMappings map[string]Mapping
	*state.Detect
}

func (b Basic) Identifier() string {
	return "basic:" + b.Name
}

func (Basic) Perform(refmap.Grapher, context.Context) error {
	return nil
}

func (b Basic) Output() string {
	return ""
}

func (e Basic) Variables() map[string]string {
	return e.Vars
}

func (b Basic) Derived() (string, string) {
	return b.SrcDerived, b.DstDerived
}

func (e *Basic) AddPosibleMapping(m Mapping) {
	e.posibleMappings[m.StartSet+"-"+m.EndSet] = m
}

func (e *Basic) Load(f io.Reader) error {
	dec := json.NewDecoder(f)

	err := dec.Decode(e)
	if err != nil {
		return err
	}

	return nil
}

func (e *Basic) Process(bb BranchBuilder, rm refmap.Mutator, ctx context.Context) error {
	if e.This == nil {
		e.This = e
	}

	rm.AddRef(ctx, e.This.Identifier(), e.This)

	err := e.ProcessState()
	if err != nil {
		return err
	}

	if e.Import {
		rootSrcDir := ctx.Value(refmap.ContextKey("source")).(string)
		metafile := filepath.Join(rootSrcDir, e.SrcDerived, "/meta.json")

		f, err := os.Open(metafile)
		if err != nil {
			return err
		}
		defer f.Close()

		eImport := &Basic{}
		err = eImport.Load(f)
		if err != nil {
			return err
		}

		if metfileWatcher, ok := ctx.Value(refmap.ContextKey("watcher")).(*fsnotify.Watcher); ok {
			metfileWatcher.Add(metafile)
		}

		if e.Directories == nil {
			e.Directories = map[string]*Directory{}
		}

		for k, v := range eImport.Directories {
			e.Directories[k] = v
			// updatePaths(d.Directories[k], d.SrcDerived)
		}

		if e.Files == nil {
			e.Files = map[string]*File{}
		}

		for k, v := range eImport.Files {
			e.Files[k] = v
			// updatePaths(e.Files[k], e.SrcDerived)
		}
	}

	if e.posibleMappings == nil {
		e.posibleMappings = map[string]Mapping{}
	}

	if e.Controls.Behaviour == nil {
		e.Controls.Behaviour = &Behaviour{}
	}
	if e.Controls.Behaviour.Filters == nil {
		e.Controls.Behaviour.Filters = filters{}
	}

	mappings := e.Controls.Mappings
	options := []string{}

	if e.Parent != nil {
		if e.Parent.Options() != "" {
			for _, option := range strings.Split(e.Parent.Options(), ",") {
				if !strings.Contains(e.Controls.Behaviour.Options, option) {
					options = append(options, option)
				}
			}
		}

		e.Controls.Mappings = append(e.Controls.Mappings, e.Parent.ControlMappings()...)

		// for i, filter := range e.Parent.Filters() {
		// 	if _, exist := e.Controls.Behaviour.Filters[i]; !exist {
		// 		e.Controls.Behaviour.Filters[i] = filter
		// 	}
		// }
	}

	if e.Controls.Behaviour.Options != "" {
		for _, option := range strings.Split(e.Controls.Behaviour.Options, ",") {
			if !strings.HasPrefix(option, "-") {
				options = append(options, option)
			}
		}
	}
	e.Controls.Behaviour.Options = strings.Join(options, ",")

	for name := range e.Files {
		e.Files[name].Name = name
		e.Files[name].Parent = e.This
		e.Files[name].Detect = state.New()
		err := e.Files[name].Process(bb, rm, ctx)
		if err != nil {
			return err
		}
	}

	for name := range e.Directories {
		e.Directories[name].Name = name
		e.Directories[name].Parent = e.This
		e.Directories[name].Detect = state.New()
		err := e.Directories[name].Process(bb, rm, ctx)
		if err != nil {
			return err
		}
	}

	for name := range e.Execs {
		e.Execs[name].Name = name
		e.Execs[name].Parent = e.This
		e.Execs[name].Detect = state.New()
		err := e.Execs[name].Process(rm, ctx)
		if err != nil {
			return err
		}
	}

	if e.Parent != nil {
		for _, m := range e.posibleMappings {
			e.Parent.AddPosibleMapping(m)
		}

		err = rm.MapRef(ctx, e.Parent.Identifier(), e.This.Identifier())
		if err != nil {
			return fmt.Errorf("mapping reference, %w", err)
		}
	}

	for _, m := range mappings {
		for _, ms := range e.posibleMappings {
			if ms.StartSet != "" && m.Start.MatchString(ms.StartSet) {
				for _, me := range e.posibleMappings {
					if me.EndSet != "" && m.End.MatchString(me.EndSet) {
						err := rm.MapRef(ctx, ms.StartSet, me.EndSet)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func (e *Basic) ProcessState() error {
	tmp := *e
	tmp.Directories = map[string]*Directory{}
	tmp.Files = map[string]*File{}
	tmp.Execs = map[string]*CLE{}
	tmp.This = nil
	tmp.Parent = nil
	tmp.posibleMappings = map[string]Mapping{}
	tmp.Detect = nil
	return e.Detect.ProcessState(fmt.Sprintf("%+v", tmp))
}
