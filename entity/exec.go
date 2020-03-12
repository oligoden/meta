package entity

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"github.com/oligoden/meta/entity/state"
)

type cle struct {
	Cmd      []string `json:"cmd"`
	Timeout  uint     `json:"timeout"`
	STDOut   *bytes.Buffer
	STDErr   *bytes.Buffer
	Parent   UpStepper `json:"-"`
	ParentID string    `json:"-"`
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

func (e *cle) Process() error {
	err := e.calculateHash()
	if err != nil {
		return err
	}
	return nil
}

func (e *cle) Perform(ctx context.Context) error {
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
	return cmd.Run()
}
