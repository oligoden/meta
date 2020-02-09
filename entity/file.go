package entity

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type file struct {
	Name      string            `json:"name"`
	Copy      bool              `json:"copy-only"`
	Source    string            `json:"source"`
	Templates map[string]string `json:"templates"`
	Parent    UpStepper         `json:"-"`
	Branch    DataBranch        `json:"-"`
	Detection
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

func (file *file) Perform(ctx context.Context) error {
	RootSrcDir := ctx.Value(ContextKey("source")).(string)
	RootDstDir := ctx.Value(ContextKey("destination")).(string)

	srcFilename := file.Name
	srcDir := ""
	dstFilename := strings.TrimSuffix(file.Name, ".tmpl")

	if file.Source != "" {
		srcFilename = filepath.Base(file.Source)
		srcDir = filepath.Dir(file.Source)
	}

	parentDS := file.Parent.(*Directory)
	defaultSrcDir := parentDS.SourcePath
	defaultDstDir := parentDS.DestinationPath
	srcFileLocation := filepath.Join(RootSrcDir, defaultSrcDir, srcDir, srcFilename)
	dstFileLocation := filepath.Join(RootDstDir, defaultDstDir, dstFilename)

	if _, err := os.Stat(dstFileLocation); err == nil {
		if !ctx.Value(ContextKey("force")).(bool) {
			return nil
		}
	} else if os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(RootDstDir, defaultDstDir), os.ModePerm)
	} else {
		return err
	}

	fmt.Println("create", dstFileLocation)
	f, err := os.Create(dstFileLocation)
	if err != nil {
		return err
	}

	if parentDS.Copy || file.Copy {
		r, err := os.Open(srcFileLocation)
		if err != nil {
			return err
		}
		defer r.Close()
		_, err = io.Copy(f, r)
		if err != nil {
			return err
		}
		err = f.Sync()
		if err != nil {
			return err
		}
	} else {
		if parentDS.Template == nil || ctx.Value(ContextKey("watching")).(bool) {
			parentDS.Template = new(Templax)
			err = parentDS.Template.Prepare(filepath.Join(RootSrcDir, defaultSrcDir))
			if err != nil {
				return err
			}
		}

		if file.Source != "" {
			if filepath.Dir(file.Source) != "." {
				err := parentDS.Template.Prepare(srcFileLocation)
				if err != nil {
					return err
				}
			}
		}

		err = parentDS.Template.FExecute(f, srcFilename, file.Branch)
		if err != nil {
			return fmt.Errorf("error executing template, %w", err)
		}
	}
	f.Close()

	return nil
}
