package state

import (
	"crypto/sha1"
	"encoding/json"
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

func NewDetect() *Detect {
	return &Detect{
		state: Added,
	}
}

func (cd Detect) Hash() string {
	return cd.hash
}

func (cd *Detect) HashOf(m interface{}) error {
	json, err := json.Marshal(m)
	if err != nil {
		return err
	}

	h := sha1.New()
	_, err = h.Write(json)
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

func (cd *Detect) State(state ...uint8) uint8 {
	if len(state) > 0 {
		cd.state = state[0]
	}
	return cd.state
}
