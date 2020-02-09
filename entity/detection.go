package entity

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"

	"github.com/oligoden/meta/mapping"
)

type Detection struct {
	hash   string
	change uint8
}

func (cd Detection) Hash() string {
	return cd.hash
}

func (cd *Detection) HashOf(m interface{}) error {
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
			cd.change = mapping.DataAdded
		} else {
			cd.change = mapping.DataUpdated
		}
		cd.hash = hash
	} else {
		cd.change = mapping.DataChecked
	}

	return nil
}

func (cd *Detection) Change(change ...uint8) uint8 {
	// 	if len(change) > 0 {
	// 		cd.change = change[0]
	// 	}
	return cd.change
}
