package entity

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
)

type file struct {
	Name      string       `json:"name"`
	Source    string       `json:"source"`
	Templates []string     `json:"templates"`
	Controls  controls     `json:"controls"`
	Template  *Templax     `json:"-"`
	Parent    ConfigReader `json:"-"`
	Branch    interface{}  `json:"-"`
	state.Detect
}

func (file *file) calculateHash() error {
	fileTemp := *file
	err := file.HashOf(fileTemp)
	if err != nil {
		return err
	}
	return nil
}

func (file file) Identifier() string {
	return "file:" + file.Source
}

func (file *file) Perform(rm refmap.Grapher, ctx context.Context) error {
	verboseValue := ctx.Value(ContextKey("verbose")).(int)
	srcFilename := filepath.Base(file.Source)

	if !file.OptionsContain(OutputBehaviour) {
		if verboseValue >= 2 {
			fmt.Println("not outputing", srcFilename)
		}
		return nil
	}

	if verboseValue >= 1 {
		fmt.Println("writing", srcFilename)
	}

	parentDS := file.Parent
	defaultSrcDir, defaultDstDir := parentDS.Derived()

	RootSrcDir := ctx.Value(ContextKey("source")).(string)
	srcDirectory := filepath.Join(RootSrcDir, defaultSrcDir)
	srcDirSpecific := filepath.Join(RootSrcDir, filepath.Dir(file.Source))
	srcFileSpecific := filepath.Join(srcDirSpecific, srcFilename)

	RootDstDir := ctx.Value(ContextKey("destination")).(string)
	dstFilename := strings.TrimSuffix(file.Name, ".tmpl")
	dstDirectory := filepath.Join(RootDstDir, defaultDstDir)
	dstFile := filepath.Join(dstDirectory, dstFilename)

	if _, err := os.Stat(dstFile); err == nil {
		if !ctx.Value(ContextKey("force")).(bool) {
			return nil
		}
	} else if os.IsNotExist(err) {
		os.MkdirAll(dstDirectory, os.ModePerm)
	} else {
		return err
	}

	contentBuf := &bytes.Buffer{}

	if file.OptionsContain(CopyBehaviour) {
		r, err := os.Open(srcFileSpecific)
		if err != nil {
			return err
		}
		defer r.Close()
		_, err = io.Copy(contentBuf, r)
		if err != nil {
			return err
		}
	} else {
		if file.Template == nil {
			file.Template = new(Templax)
		}

		err := file.Template.Prepare(srcFileSpecific)
		if err != nil {
			return err
		}

		if file.OptionsContain("parse-dir") {
			err = file.Template.Prepare(srcDirectory)
			if err != nil {
				if !strings.Contains(err.Error(), "template: pattern matches no files") {
					return err
				}
				log.Println(err)
			}
		}

		for _, t := range rm.ParentFiles(file.Identifier()) {
			f := strings.TrimPrefix(t, "file:")
			err := file.Template.Prepare(filepath.Join(RootSrcDir, f))
			if err != nil {
				return err
			}
		}

		for _, template := range file.Templates {
			template = filepath.Join(strings.Split(template, "/")...)

			err := file.Template.Prepare(filepath.Join(RootSrcDir, template))
			if err != nil {
				return err
			}

			for _, t := range rm.ParentFiles("file:" + template) {
				f := strings.TrimPrefix(t, "file:")
				err := file.Template.Prepare(filepath.Join(RootSrcDir, f))
				if err != nil {
					return err
				}
			}
		}

		err = file.Template.FExecute(contentBuf, srcFilename, file.Branch)
		if err != nil {
			return fmt.Errorf("error executing template, %w", err)
		}
	}

	outputBuf := &bytes.Buffer{}

	if file.ContainsFilter("comment-filter") {
		err := lineFilter(contentBuf, outputBuf)
		if err != nil {
			return fmt.Errorf("error with line control, %w", err)
		}
	} else {
		contentBuf.WriteTo(outputBuf)
	}

	f, err := os.OpenFile(dstFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = outputBuf.WriteTo(f)
	if err != nil {
		return fmt.Errorf("error writing to file, %w", err)
	}

	if err := f.Close(); err != nil {
		log.Println("error closing file", err)
	}

	return nil
}

func (f file) OptionsContain(b string) bool {
	return f.Parent.OptionsContain(b) || f.Controls.Behaviour.Options == b
}

func (f file) ContainsFilter(filter string) bool {
	if _, has := f.Controls.Behaviour.Filters[filter]; has {
		return true
	}
	return false
}

func (f file) Output() string {
	return ""
}

func lineFilter(r, w *bytes.Buffer) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "//---") {
			for scanner.Scan() {
				line = scanner.Text()
				if strings.HasPrefix(strings.TrimSpace(line), "//end") {
					break
				}
			}
		} else if strings.Contains(strings.TrimSpace(line), "//-") {
			// skip line
		} else if strings.Contains(strings.TrimSpace(line), "//>") {
			fmt.Fprintln(w, strings.Replace(line, "//>", "//-", 1))
		} else if strings.HasPrefix(strings.TrimSpace(line), "//+++") {
			for scanner.Scan() {
				line = scanner.Text()
				if strings.HasPrefix(strings.TrimSpace(line), "//end") {
					break
				}
				if strings.HasPrefix(strings.TrimSpace(line), "//") {
					line = strings.Replace(line, "//", "", 1)
				}
				fmt.Fprintln(w, line)
			}
		} else {
			fmt.Fprintln(w, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(w, "reading standard input:", err)
	}

	return nil
}
