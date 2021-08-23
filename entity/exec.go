package entity

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/oligoden/meta/entity/state"
	"github.com/oligoden/meta/refmap"
)

type cle struct {
	Name        string
	Cmd         []string          `json:"cmd"`
	Timeout     uint              `json:"timeout"`
	Environment map[string]string `json:"env"`
	STDOut      *bytes.Buffer
	STDErr      *bytes.Buffer
	Parent      ConfigReader `json:"-"`
	state.Detect
}

func (exec *cle) calculateHash() error {
	err := exec.HashOf()
	if err != nil {
		return err
	}
	return nil
}

func (e cle) Identifier() string {
	return "exec:" + e.Name
}

func (e cle) Output() string {
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

func (e cle) Derived() (string, string) {
	return "", ""
}

func (e *cle) Process(rm refmap.Mutator, ctx context.Context) error {
	err := e.HashOf()
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

func (e *cle) Perform(rm refmap.Grapher, ctx context.Context) error {
	RootDstDir := ctx.Value(refmap.ContextKey("destination")).(string)

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
	cmd.Dir = RootDstDir
	cmd.Stdout = e.STDOut
	cmd.Stderr = e.STDErr
	for k, v := range e.Environment {
		cmd.Env = append(os.Environ(), fmt.Sprintf(`%s=%s`, k, v))
	}
	return cmd.Run()
}
