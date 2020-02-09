package entity

import (
	"path/filepath"
	"strings"

	"github.com/oligoden/meta/mapping"
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

func (d *Directory) Process(bb func(BranchSetter) (UpStepper, error), m mapping.Mutator) error {
	err := d.calculateHash()
	if err != nil {
		return err
	}

	d.SourcePath = path(d.SourcePath, d.Source)
	d.DestinationPath = path(d.DestinationPath, d.Destination)

	for name, dir := range d.Directories {
		dir.Parent = d
		dir.SourcePath = filepath.Join(d.SourcePath, name)
		dir.DestinationPath = filepath.Join(d.DestinationPath, name)
		dir.Name = name
		dir.Process(bb, m)
	}

	for name, file := range d.Files {
		file.Name = name
		file.Parent = d

		err := file.calculateHash()
		if err != nil {
			return err
		}

		_, err = bb(file)
		if err != nil {
			return err
		}

		filename := name
		// 	if file.Source != "" {
		// 		filename = file.Source
		// 	}

		if m != nil {
			source := filepath.Join(d.SourcePath, filename)
			destination := filepath.Join(d.DestinationPath, name)
			m.Write(source, destination, file)
		}
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
