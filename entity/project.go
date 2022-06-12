package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
)

func NewProject() *Project {
	e := &Project{}
	e.This = e
	e.Controls = NewControls()
	e.Detect = state.New()
	return e
}

type Project struct {
	Testing      bool       `json:"testing"`
	Environment  string     `json:"environment"`
	Repository   Repository `json:"repo"`
	WorkLocation string     `json:"work-location"`
	DestLocation string     `json:"dest-location"`
	oldName      string
	Basic
}

type Repository struct {
}

func (p Project) Identifier() string {
	return "prj:" + p.Name
}

func (e *Project) Load(f io.Reader) error {
	dec := json.NewDecoder(f)

	err := dec.Decode(e)
	if err != nil {
		return err
	}

	return nil
}

func (e *Project) LoadFile(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return fmt.Errorf("opening file, %w", err)
	}

	err = e.Load(f)
	if err != nil {
		return fmt.Errorf("loading file, %w", err)
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("closing file, %w", err)
	}

	fmt.Println("loaded config file")
	return nil
}

func (e *Project) Process(bb BranchBuilder, rm refmap.Mutator, ctx context.Context) error {
	// Check if name changed
	if e.oldName != "" && e.oldName != e.Name {
		if e.oldName == "-" {
			if e.Name != "" {
				rm.RenameRef(ctx, "prj:", e.Identifier())
			}
		} else {
			rm.RenameRef(ctx, "prj:"+e.oldName, e.Identifier())
		}
	}

	e.oldName = e.Name
	if e.oldName == "" {
		e.oldName = "-"
	}

	err := e.Basic.Process(bb, rm, ctx)
	if err != nil {
		return err
	}

	return nil
}
