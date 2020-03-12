package entity

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/oligoden/meta/entity/state"
)

type file struct {
	Name      string            `json:"name"`
	Copy      bool              `json:"copy-only"`
	DontWrite bool              `json:"dont-write"`
	Source    string            `json:"source"`
	IgnoreDS  bool              `json:"ignore-default"`
	Templates map[string]string `json:"templates"`
	Parent    UpStepper         `json:"-"`
	ParentID  string            `json:"-"`
	Branch    DataBranch        `json:"-"`
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

func (file *file) Perform(ctx context.Context) error {
	RootSrcDir := ctx.Value(ContextKey("source")).(string)
	RootDstDir := ctx.Value(ContextKey("destination")).(string)

	srcFilename := filepath.Base(file.Source)
	dstFilename := strings.TrimSuffix(file.Name, ".tmpl")

	parentDS := file.Parent.(*Directory)
	defaultSrcDir := parentDS.SourcePath
	defaultDstDir := parentDS.DestinationPath
	srcDirLocation := filepath.Join(RootSrcDir, defaultSrcDir)
	srcDirFullLocation := filepath.Join(RootSrcDir, filepath.Dir(file.Source))
	dstDirLocation := filepath.Join(RootDstDir, defaultDstDir)
	srcFileLocation := filepath.Join(srcDirFullLocation, srcFilename)
	dstFileLocation := filepath.Join(dstDirLocation, dstFilename)

	// if _, err := os.Stat(dstFileLocation); err == nil {
	// 	if !ctx.Value(ContextKey("force")).(bool) {
	// 		return nil
	// 	}
	// } else if os.IsNotExist(err) {
	os.MkdirAll(dstDirLocation, os.ModePerm)
	// } else {
	// 	return err
	// }

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
			err = parentDS.Template.Prepare(srcDirLocation)
			if err != nil {
				return err
			}
		}

		if srcDirLocation != srcDirFullLocation {
			err := parentDS.Template.Prepare(srcFileLocation)
			if err != nil {
				return err
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
