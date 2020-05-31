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
)

type file struct {
	Name      string     `json:"name"`
	Copy      bool       `json:"copy-only"`
	DontWrite bool       `json:"dont-write"`
	Source    string     `json:"source"`
	IgnoreDS  bool       `json:"ignore-default"`
	Templates []string   `json:"templates"`
	Parent    UpStepper  `json:"-"`
	ParentID  string     `json:"-"`
	Branch    DataBranch `json:"-"`
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

func (file *file) Perform(ctx context.Context) error {
	verboseValue := ctx.Value(ContextKey("verbose")).(int)
	RootSrcDir := ctx.Value(ContextKey("source")).(string)
	RootDstDir := ctx.Value(ContextKey("destination")).(string)

	srcFilename := filepath.Base(file.Source)
	if verboseValue >= 1 {
		fmt.Println("writing", srcFilename)
	}
	dstFilename := strings.TrimSuffix(file.Name, ".tmpl")

	parentDS := file.Parent.(*Directory)
	defaultSrcDir := parentDS.SourcePath
	defaultDstDir := parentDS.DestinationPath
	srcDirLocation := filepath.Join(RootSrcDir, defaultSrcDir)
	srcDirFullLocation := filepath.Join(RootSrcDir, filepath.Dir(file.Source))
	dstDirLocation := filepath.Join(RootDstDir, defaultDstDir)
	srcFileLocation := filepath.Join(srcDirFullLocation, srcFilename)
	dstFileLocation := filepath.Join(dstDirLocation, dstFilename)

	if _, err := os.Stat(dstFileLocation); err == nil {
		if !ctx.Value(ContextKey("force")).(bool) {
			return nil
		}
	} else if os.IsNotExist(err) {
		os.MkdirAll(dstDirLocation, os.ModePerm)
	} else {
		return err
	}

	f, err := os.OpenFile(dstFileLocation, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	contentBuf := &bytes.Buffer{}
	if parentDS.Copy || file.Copy {
		r, err := os.Open(srcFileLocation)
		if err != nil {
			return err
		}
		defer r.Close()
		_, err = io.Copy(contentBuf, r)
		if err != nil {
			return err
		}
	} else {
		if parentDS.Template == nil || ctx.Value(ContextKey("watching")).(bool) {
			parentDS.Template = new(Templax)
			err = parentDS.Template.Prepare(srcDirLocation)
			if err != nil {
				if !strings.Contains(err.Error(), "template: pattern matches no files") {
					return err
				}
				log.Println(err)
			}
		}

		if srcDirLocation != srcDirFullLocation {
			err := parentDS.Template.Prepare(srcFileLocation)
			if err != nil {
				return err
			}
		}

		for _, template := range file.Templates {
			err := parentDS.Template.Prepare(filepath.Join(RootSrcDir, template))
			if err != nil {
				return err
			}
		}

		err = parentDS.Template.FExecute(contentBuf, srcFilename, file.Branch)
		if err != nil {
			return fmt.Errorf("error executing template, %w", err)
		}
	}

	commentFilteredBuf := &bytes.Buffer{}
	err = lineFilter(contentBuf, commentFilteredBuf)
	if err != nil {
		return fmt.Errorf("error with line control, %w", err)
	}

	_, err = commentFilteredBuf.WriteTo(f)
	if err != nil {
		return fmt.Errorf("error writing to file, %w", err)
	}

	if err := f.Close(); err != nil {
		log.Println("error closing file", err)
	}

	return nil
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
