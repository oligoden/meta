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
	Parent      Identifier `json:"-"`
	state.Detect
}

func (exec *cle) calculateHash() error {
	execTemp := *exec
	err := exec.HashOf(execTemp)
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

// func (e *cle) Process() error {
// 	err := e.calculateHash()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (e *cle) Perform(rm refmap.Grapher, ctx context.Context) error {
	RootDstDir := ctx.Value(ContextKey("destination")).(string)

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
