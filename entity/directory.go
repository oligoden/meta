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

	// Settings can contain:
	// - "copy-only" to only copy file
	// - "parse-dir" to parse all templates in directory
	// - "comment-filter" to apply comment line filter
	// - "no-output" to skip file output
	Settings string   `json:"settings"`
	LinkTo   []string `json:"linkto"`
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
	d.SourcePath = path(d.SourcePath, d.Source)
	d.DestinationPath = path(d.DestinationPath, d.Destination)

	if d.Edges == nil {
		d.Edges = []Edge{}
	}

	if d.LinkTo == nil {
		d.LinkTo = []string{}
	}

	err := d.calculateHash()
	if err != nil {
		return err
	}

	refName := fmt.Sprintf("dir:%s:%s", d.SourcePath, d.Name)
	m.AddRef(refName, d)
	m.MapRef(d.ParentID, refName)

	for name, e := range d.Execs {
		e.Name = name
		e.Parent = d
		e.ParentID = refName
		err := e.calculateHash()
		if err != nil {
			return err
		}
		m.AddRef("exec:"+name, e)
		m.MapRef(refName, "exec:"+name)
		d.LinkTo = append(d.LinkTo, "exec:"+name)
	}

	for name, dir := range d.Directories {
		dir.Parent = d
		dir.ParentID = refName
		dir.SourcePath = filepath.Join(d.SourcePath, name)
		dir.DestinationPath = filepath.Join(d.DestinationPath, name)
		dir.Name = name
		dir.LinkTo = d.LinkTo
		dir.Edges = d.Edges
		dir.Process(bb, m)
		d.Edges = dir.Edges
	}

	for name, file := range d.Files {
		file.Name = name
		file.Parent = d
		file.ParentID = refName

		err := file.calculateHash()
		if err != nil {
			return err
		}

		_, err = bb(file)
		if err != nil {
			return err
		}

		if file.Source == "" {
			file.Source = filepath.Join(d.SourcePath, name)
		} else if strings.HasPrefix(file.Source, "./") {
			file.Source = filepath.Join(d.SourcePath, file.Source)
		}

		m.AddRef("file:"+file.Source, file)
		m.MapRef(file.ParentID, "file:"+file.Source)

		for _, t := range file.Templates {
			fmt.Println("adding", t, file.Source)
			d.Edges = append(d.Edges, Edge{
				Start: "file:" + t,
				End:   "file:" + file.Source,
			})
		}

		for _, lt := range d.LinkTo {
			m.MapRef("file:"+file.Source, lt)
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
