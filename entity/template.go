package entity

import (
	"fmt"
	"io"
	"os"
	"text/template"
)

type Templax struct {
	all *template.Template
}

func (t *Templax) FExecute(f io.Writer, fn string, data interface{}) error {
	if t.all == nil {
		return fmt.Errorf("no templates parsed yet")
	}

	selected := t.all.Lookup(fn)
	if selected == nil {
		return fmt.Errorf("file %s not prepared", fn)
	}

	err := selected.Execute(f, data)
	if err != nil {
		fmt.Println("execute error")
		return err
	}
	return nil
}

func (t *Templax) Prepare(fp string) error {
	stat, err := os.Stat(fp)
	if err != nil {
		return err
	}

	var tmpl *template.Template

	if stat.IsDir() {
		pt := fmt.Sprintf("%[1]s/*.*", fp)
		if t.all == nil {
			tmpl, err = template.ParseGlob(pt)
		} else {
			tmpl, err = t.all.ParseGlob(pt)
		}
	} else {
		if t.all == nil {
			tmpl, err = template.ParseFiles(fp)
		} else {
			tmpl, err = t.all.ParseFiles(fp)
		}
	}

	if err != nil {
		return err
	}
	t.all = tmpl
	return nil
}
