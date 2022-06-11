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

func New(hs ...string) *Detect {
	h := ""
	if len(hs) > 0 {
		h = hs[0]
	}

	if h == "" {
		return &Detect{
			state: Added,
		}
	} else {
		return &Detect{
			hash: h,
		}
	}
}

func (cd Detect) Hash() string {
	return cd.hash
}

func (cd *Detect) ProcessState(s string) error {
	if cd.state != Stable {
		if cd.state != Added || cd.hash != "" {
			return nil
		}
	}

	data := []byte(s)
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

func (cd *Detect) RemoveState() {
	cd.state = Remove
}
