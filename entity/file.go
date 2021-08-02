package entity

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
)

type File struct {
	Name     string             `json:"name"`
	Source   string             `json:"source"`
	Controls controls           `json:"controls"`
	Template *template.Template `json:"-"`
	Parent   ConfigReader       `json:"-"`
	Branch   interface{}        `json:"-"`
	state.Detect
}

func (file File) Identifier() string {
	return "file:" + file.Source
}

func (e *File) Process(bb BranchBuilder, rm refmap.Mutator, ctx context.Context) error {
	if e.Controls.Behaviour == nil {
		e.Controls.Behaviour = &behaviour{}
	}
	if e.Controls.Behaviour.Filters == nil {
		e.Controls.Behaviour.Filters = filters{}
	}

	options := []string{}
	for _, option := range strings.Split(e.Parent.Options(), ",") {
		if !strings.Contains(e.Controls.Behaviour.Options, option) {
			options = append(options, option)
		}
	}

	for _, option := range strings.Split(e.Controls.Behaviour.Options, ",") {
		if !strings.HasPrefix(option, "-") {
			options = append(options, option)
		}
	}
	e.Controls.Behaviour.Options = strings.Join(options, ",")

	for i, filter := range e.Parent.Filters() {
		if _, exist := e.Controls.Behaviour.Filters[i]; !exist {
			e.Controls.Behaviour.Filters[i] = filter
		}
	}

	err := e.HashOf()
	if err != nil {
		return err
	}

	_, err = bb.Build(e)
	if err != nil {
		return fmt.Errorf("building branch, %w", err)
	}
	e.Branch = bb

	srcDerived, _ := e.Parent.Derived()
	if e.Source == "" {
		e.Source = filepath.Join(srcDerived, e.Name)
	} //else if strings.HasPrefix(e.Source, "./") {
	// 	e.Source = filepath.Join(parent.SrcDerived, e.Source)
	// }

	for _, m := range e.Parent.ControlMappings() {
		matchStart := m.Start.MatchString(e.Identifier())
		matchEnd := m.End.MatchString(e.Identifier())
		if matchStart && matchEnd {
			return fmt.Errorf("directory matches start and end reference")
		}
		if matchStart {
			e.Parent.AddPosibleMapping(Mapping{
				StartSet:   e.Identifier(),
				End:        m.End,
				Recurrence: m.Recurrence,
			})
		}
		if matchEnd {
			e.Parent.AddPosibleMapping(Mapping{
				Start:      m.Start,
				EndSet:     e.Identifier(),
				Recurrence: m.Recurrence,
			})
		}
	}

	rm.AddRef(ctx, "file:"+e.Source, e)
	err = rm.MapRef(ctx, e.Parent.Identifier(), "file:"+e.Source)
	if err != nil {
		return err
	}

	return nil
}

func (file *File) Perform(rm refmap.Grapher, ctx context.Context) error {
	verboseValue := ctx.Value(refmap.ContextKey("verbose")).(int)
	srcFilename := filepath.Base(file.Source)

	if !strings.Contains(file.Controls.Behaviour.Options, "output") {
		if verboseValue >= 2 {
			fmt.Println("not outputing", srcFilename)
		}
		return nil
	}

	if verboseValue >= 3 {
		fmt.Println("writing", srcFilename)
	}

	_, defaultDstDir := file.Parent.Derived()
	defaultSrcDir, defaultDstDir := file.Parent.Derived()

	RootSrcDir := ctx.Value(refmap.ContextKey("source")).(string)
	srcDirectory := filepath.Join(RootSrcDir, defaultSrcDir)
	srcFile := filepath.Join(srcDirectory, srcFilename)
	// srcDirSpecific := filepath.Join(RootSrcDir, filepath.Dir(file.Source))
	// srcFileSpecific := filepath.Join(srcDirSpecific, srcFilename)

	RootDstDir := ctx.Value(refmap.ContextKey("destination")).(string)
	dstFilename := strings.TrimSuffix(file.Name, ".tmpl")
	dstDirectory := filepath.Join(RootDstDir, defaultDstDir)
	dstFile := filepath.Join(dstDirectory, dstFilename)

	_, err := os.Stat(dstFile)
	if err == nil {
		// 	if !ctx.Value(ContextKey("force")).(bool) {
		// 		return nil
		// 	}
	} else if os.IsNotExist(err) {
		os.MkdirAll(dstDirectory, os.ModePerm)
	} else {
		return err
	}

	contentBuf := &bytes.Buffer{}

	if strings.Contains(file.Controls.Behaviour.Options, "copy") {
		r, err := os.Open(srcFile)
		if err != nil {
			return err
		}
		defer r.Close()
		_, err = io.Copy(contentBuf, r)
		if err != nil {
			return err
		}
	} else {
		fileContent, err := ioutil.ReadFile(srcFile)
		if err != nil {
			return err
		}

		tmpl, err := template.New(srcFile).Parse(string(fileContent))
		if err != nil {
			return err
		}

		for _, t := range rm.ParentFiles(file.Identifier()) {
			filename := strings.TrimPrefix(t, "file:")
			filename = filepath.Join(RootSrcDir, filename)

			fileContent, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}

			tmpl, err = tmpl.New(filename).Parse(string(fileContent))
			if err != nil {
				return err
			}
		}

		err = tmpl.Lookup(srcFile).Execute(contentBuf, file.Branch)
		if err != nil {
			return fmt.Errorf("error executing template, %w", err)
		}
	}

	outputBuf := &bytes.Buffer{}

	if file.ContainsFilter("comment") {
		err := commentFilter(contentBuf, outputBuf)
		if err != nil {
			return fmt.Errorf("error with comment filter, %w", err)
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

func (e File) ContainsFilter(filter string) bool {
	if _, has := e.Controls.Behaviour.Filters[filter]; has {
		return true
	}
	return false
}

func (f File) Output() string {
	return ""
}

func commentFilter(r, w *bytes.Buffer) error {
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
