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
	Name      string   `json:"name"`
	Source    string   `json:"source"`
	Templates []string `json:"templates"`

	// Settings can contain:
	// - "copy-only" to only copy file
	// - "parse-dir" to parse all templates in directory
	// - "comment-filter" to apply comment line filter
	// - "no-output" to skip file output
	Settings string     `json:"settings"`
	Template *Templax   `json:"-"`
	Parent   UpStepper  `json:"-"`
	ParentID string     `json:"-"`
	Branch   DataBranch `json:"-"`
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

func (file *file) SetBranch(b ...DataBranch) DataBranch {
	if len(b) > 0 {
		file.Branch = b[0]
	}
	return file.Branch
}

func (file file) Identifier() string {
	return file.Source
}

func (file *file) Perform(rm refmap.Grapher, ctx context.Context) error {
	verboseValue := ctx.Value(ContextKey("verbose")).(int)
	parentDS := file.Parent.(*Directory)

	if has(parentDS, file, "no-output") {
		return nil
	}

	RootSrcDir := ctx.Value(ContextKey("source")).(string)
	srcFilename := filepath.Base(file.Source)
	if verboseValue >= 1 {
		fmt.Println("writing", srcFilename)
	}
	defaultSrcDir := parentDS.SourcePath
	srcDirectory := filepath.Join(RootSrcDir, defaultSrcDir)
	srcDirSpecific := filepath.Join(RootSrcDir, filepath.Dir(file.Source))
	srcFileSpecific := filepath.Join(srcDirSpecific, srcFilename)

	RootDstDir := ctx.Value(ContextKey("destination")).(string)
	dstFilename := strings.TrimSuffix(file.Name, ".tmpl")
	defaultDstDir := parentDS.DestinationPath
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

	if has(parentDS, file, "copy-only") {
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

		if has(parentDS, file, "parse-dir") {
			err = file.Template.Prepare(srcDirectory)
			if err != nil {
				if !strings.Contains(err.Error(), "template: pattern matches no files") {
					return err
				}
				log.Println(err)
			}
		}

		for _, template := range file.Templates {
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

	if has(parentDS, file, "comment-filter") {
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

func has(d *Directory, f *file, s string) bool {
	return strings.Contains(d.Settings, s) || strings.Contains(f.Settings, s)
}

func lineFilter(r, w *bytes.Buffer) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "//xxx") {
			for scanner.Scan() {
				line = scanner.Text()
				if strings.HasPrefix(strings.TrimSpace(line), "//end") {
					break
				}
			}
		} else if strings.HasPrefix(strings.TrimSpace(line), "//xx") {
			// skip line
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
