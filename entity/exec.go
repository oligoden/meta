package entity

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
)

type CLE struct {
	Name    string
	Cmd     []string          `json:"cmd"`
	Timeout uint              `json:"timeout"`
	Env     map[string]string `json:"env"`
	Dir     string            `json:"dir"`
	STDOut  *bytes.Buffer
	STDErr  *bytes.Buffer
	Parent  ConfigReader `json:"-"`
	*state.Detect
}

func (exec *CLE) calculateHash() error {
	err := exec.ProcessState()
	if err != nil {
		return err
	}
	return nil
}

func (e CLE) Identifier() string {
	return "exec:" + e.Name
}

func (e CLE) Output() string {
	output := fmt.Sprintf("action %s was run", e.Name)
	if e.STDOut.String() != "" {
		output += "\nstdout: " + e.STDOut.String()
		e.STDOut.Reset()
	}
	if e.STDErr.String() != "" {
		output += "\nstderr: " + e.STDErr.String()
		e.STDErr.Reset()
	}
	return output
}

func (e CLE) Derived() (string, string) {
	return "", ""
}

func (e *CLE) Process(rm refmap.Mutator, ctx context.Context) error {
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

	rm.AddRef(ctx, e.Identifier(), e)
	err = rm.MapRef(ctx, e.Parent.Identifier(), e.Identifier())
	if err != nil {
		return err
	}

	return nil
}

func (e *CLE) Perform(rm refmap.Grapher, ctx context.Context) error {
	RootSrcDir := ctx.Value(refmap.ContextKey("orig")).(string)

	if e.Timeout == 0 {
		e.Timeout = 500
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(e.Timeout)*time.Millisecond)
	defer cancel()

	if e.STDErr == nil {
		e.STDErr = &bytes.Buffer{}
	}
	if e.STDOut == nil {
		e.STDOut = &bytes.Buffer{}
	}

	cmd := exec.CommandContext(ctx, e.Cmd[0], e.Cmd[1:]...)
	cmd.Dir = filepath.Join(RootSrcDir, e.Dir)
	cmd.Stdout = e.STDOut
	cmd.Stderr = e.STDErr
	for k, v := range e.Env {
		cmd.Env = append(os.Environ(), fmt.Sprintf(`%s=%s`, k, v))
	}
	return cmd.Run()
}

func (e *CLE) ProcessState() error {
	return e.Detect.ProcessState("")
}
