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
	Vars     map[string]string  `json:"vars"`
	Opts     string             `json:"options"`
	Flts     filters            `json:"filters"`
	Mpns     []*Mapping         `json:"mappings"`
	Template *template.Template `json:"-"`
	Parent   ConfigReader       `json:"-"`
	Branch   BranchBuilder      `json:"-"`
	*state.Detect
}

func (file File) Identifier() string {
	return "file:" + file.Source
}

func (e *File) Process(bb BranchBuilder, rm refmap.Mutator, ctx context.Context) error {
	if e.Flts == nil {
		e.Flts = filters{}
	}

	options := []string{}
	for _, option := range strings.Split(e.Parent.Options(), ",") {
		if !strings.Contains(e.Opts, option) {
			options = append(options, option)
		}
	}

	for _, option := range strings.Split(e.Opts, ",") {
		if !strings.HasPrefix(option, "-") {
			options = append(options, option)
		}
	}
	e.Opts = strings.Join(options, ",")

	for i, filter := range e.Parent.Filters() {
		if _, exist := e.Flts[i]; !exist {
			e.Flts[i] = filter
		}
	}

	if e.Vars == nil {
		e.Vars = map[string]string{}
	}
	for k, v := range e.Parent.Variables() {
		if _, ok := e.Vars[k]; !ok {
			e.Vars[k] = v
		}
	}

	hash := ""
	nodes := rm.Nodes("", e.Identifier())
	if len(nodes) > 0 {
		hash = nodes[0].Hash()
	}
	e.Detect = state.New(hash)

	err := e.ProcessState()
	if err != nil {
		return err
	}

	e.Branch = bb.Clone()
	_, err = e.Branch.Build(e)
	if err != nil {
		return fmt.Errorf("building branch, %w", err)
	}

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
		return fmt.Errorf("mapping nodes, %w", err)
	}

	return nil
}

func (file *File) Perform(rm refmap.Grapher, ctx context.Context) error {
	verboseValue := ctx.Value(refmap.ContextKey("verbose")).(int)
	srcFilename := filepath.Base(file.Source)

	if !strings.Contains(file.Opts, "output") {
		if verboseValue >= 2 {
			fmt.Println("not outputing", srcFilename)
		}
		return nil
	}

	if verboseValue >= 3 {
		fmt.Println("writing", srcFilename)
	}

	defaultSrcDir, defaultDstDir := "", ""
	if file.Parent != nil {
		defaultSrcDir, defaultDstDir = file.Parent.Derived()
	}

	RootSrcDir := ctx.Value(refmap.ContextKey("orig")).(string)
	srcDirectory := filepath.Join(RootSrcDir, defaultSrcDir)
	srcFile := filepath.Join(srcDirectory, srcFilename)
	// srcDirSpecific := filepath.Join(RootSrcDir, filepath.Dir(file.Source))
	// srcFileSpecific := filepath.Join(srcDirSpecific, srcFilename)

	RootDstDir := ctx.Value(refmap.ContextKey("dest")).(string)
	dstFilename := strings.TrimSuffix(file.Name, ".tmpl")
	dstDirectory := filepath.Join(RootDstDir, defaultDstDir)
	dstFile := filepath.Join(dstDirectory, dstFilename)

	_, err := os.Stat(dstFile)
	if err == nil {
		if nd := rm.Nodes("", file.Identifier()); len(nd) > 0 {
			if nd[0].State() == state.Remove {
				if verboseValue >= 3 {
					fmt.Println(dstFile, "set for removal, deleting")
				}

				err := os.Remove(dstFile)
				if err != nil {
					fmt.Println("error deleting file", dstFile)
				}
				return nil
			}
		}
	} else if os.IsNotExist(err) {
		os.MkdirAll(dstDirectory, os.ModePerm)
	} else {
		return fmt.Errorf("stating destination file %s -> %w", dstFile, err)
	}

	contentBuf := &bytes.Buffer{}

	if strings.Contains(file.Opts, "copy") {
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

		tmpl, err := template.New(srcFile).
			Option("missingkey=error").
			Parse(string(fileContent))
		if err != nil {
			return err
		}

		if rm != nil {
			for _, t := range rm.ParentFiles(file.Identifier()) {
				filename := strings.TrimPrefix(t, "file:")
				filename = filepath.Join(RootSrcDir, filename)

				fileContent, err := ioutil.ReadFile(filename)
				if err != nil {
					return err
				}

				tmpl, err = tmpl.New(filename).
					Option("missingkey=error").
					Parse(string(fileContent))
				if err != nil {
					return err
				}
			}
		}

		err = tmpl.Lookup(srcFile).Execute(contentBuf, file.Branch)
		if err != nil {
			return fmt.Errorf("error executing template -> %w", err)
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
		return fmt.Errorf("opening destination file %s for writing -> %w", dstFile, err)
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
	if _, has := e.Flts[filter]; has {
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
		} else if strings.Contains(line, "//+") {
			fmt.Fprintln(w, strings.Replace(line, "//+", "", 1))
		} else {
			fmt.Fprintln(w, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(w, "reading standard input:", err)
	}

	return nil
}

func (e File) ProcessState() error {
	return e.Detect.ProcessState("")
}
