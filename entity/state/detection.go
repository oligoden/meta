package state

import (
	"crypto/sha1"
	"fmt"
)

const (
	Stable uint8 = iota
	Checked
	Updated
	Added
	Remove
)

type Detect struct {
	hash  string
	state uint8
}

func New() *Detect {
	return &Detect{
		state: Added,
	}
}

func (cd Detect) Hash() string {
	return cd.hash
}

func (cd *Detect) ProcessState(e string) error {
	data := []byte(e)
	h := sha1.New()
	_, err := h.Write(data)
	if err != nil {
		return err
	}

	hash := fmt.Sprintf("%x", h.Sum(nil))
	if hash != cd.hash {
		if cd.hash == "" {
			cd.state = Added
		} else {
			cd.state = Updated
		}
		cd.hash = hash
	} else {
		cd.state = Checked
	}

	return nil
}

func (cd *Detect) State() uint8 {
	return cd.state
}

func (cd *Detect) ClearState() {
	cd.state = Stable
}

func (cd *Detect) FlagState() {
	cd.state = Updated
}
