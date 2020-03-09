package entity

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/oligoden/meta/refmap"
)

type Directory struct {
	Source          string           `json:"from"`
	Destination     string           `json:"dest"`
	Files           map[string]*file `json:"files"`
	SourcePath      string           `json:"-"`
	DestinationPath string           `json:"-"`
	Template        *Templax         `json:"-"`
	Copy            bool             `json:"copy-only"`
	Basic
}

func (d *Directory) calculateHash() error {
	dirTemp := *d
	dirTemp.Directories = nil
	dirTemp.Files = nil
	err := d.HashOf(dirTemp)
	if err != nil {
		return err
	}
	return nil
}

func (d *Directory) Process(bb func(BranchSetter) (UpStepper, error), m refmap.Mutator) error {
	err := d.calculateHash()
	if err != nil {
		return err
	}

	d.SourcePath = path(d.SourcePath, d.Source)
	d.DestinationPath = path(d.DestinationPath, d.Destination)

	m.AddRef("dir:"+d.DestinationPath, d)
	m.MapRef(d.ParentID, "dir:"+d.DestinationPath)

	for name, dir := range d.Directories {
		dir.Parent = d
		dir.ParentID = "dir:" + d.DestinationPath
		dir.SourcePath = filepath.Join(d.SourcePath, name)
		dir.DestinationPath = filepath.Join(d.DestinationPath, name)
		dir.Name = name
		dir.Process(bb, m)
	}

	for name, file := range d.Files {
		file.Name = name
		file.Parent = d
		file.ParentID = "dir:" + d.DestinationPath

		err := file.calculateHash()
		if err != nil {
			return err
		}

		_, err = bb(file)
		if err != nil {
			fmt.Println("error", err)
			return err
		}

		if file.Source == "" {
			file.Source = filepath.Join(d.SourcePath, name)
		} else if strings.HasPrefix(file.Source, ".") {
			file.Source = filepath.Join(d.SourcePath, file.Source)
		}

		destination := filepath.Join(d.DestinationPath, name)
		m.AddRef("file:"+destination, file)
		m.MapRef(file.ParentID, "file:"+destination)
	}
	return nil
}

func path(path, modify string) string {
	if strings.HasPrefix(modify, ".") {
		return filepath.Join(filepath.Dir(path), modify)
	}
	if strings.HasPrefix(modify, "/") {
		return strings.TrimPrefix(modify, "/")
	}
	return filepath.Join(path, modify)
}
