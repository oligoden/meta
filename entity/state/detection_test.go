package state_test

import (
	"fmt"
	"testing"

	"github.com/oligoden/meta/entity/state"
	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	d := testEntity{
		Name:   "a",
		Detect: state.New(),
	}

	assert := assert.New(t)
	if assert.NoError(d.ProcessState()) {
		assert.NotEmpty(d.Hash())
		assert.Equal(state.Added, d.State())
	}
	d.ClearState()

	hash1 := d.Hash()

	if assert.NoError(d.ProcessState()) {
		assert.Equal(state.Checked, d.State())
	}
	d.ClearState()

	d.Name = "b"
	if assert.NoError(d.ProcessState()) {
		assert.Equal(state.Updated, d.State())
		assert.NotEqual(hash1, d.Hash())
	}
	d.ClearState()

	if assert.NoError(d.ProcessState()) {
		assert.Equal(state.Checked, d.State())
	}
	d.ClearState()

	d.FlagState()
	assert.Equal(state.Updated, d.State())

	d.RemoveState()
	assert.Equal(state.Remove, d.State())
}

type testEntity struct {
	Name string
	*state.Detect
}

func (e testEntity) ProcessState() error {
	tmp := e
	tmp.Detect = nil
	return e.Detect.ProcessState(fmt.Sprintf("%+v", tmp))
}
