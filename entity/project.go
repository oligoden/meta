package entity

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/oligoden/meta/entity/state"
)

func NewProject() *Project {
	e := &Project{}
	e.This = e
	e.Detect = state.New()
	return e
}

type Project struct {
	Testing      bool       `json:"testing"`
	Environment  string     `json:"environment"`
	Repository   Repository `json:"repo"`
	WorkLocation string     `json:"work-location"`
	DestLocation string     `json:"dest-location"`
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
