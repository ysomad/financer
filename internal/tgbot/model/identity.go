package model

import (
	"errors"
)

type Identity struct {
	ID          string
	TGUID       int64
	AccessToken string
	Currency    string
}

func (id Identity) Validate() error {
	if id.ID == "" {
		return errors.New("id is required")
	}
	if id.TGUID <= 0 {
		return errors.New("telegram uid is required")
	}
	if id.AccessToken == "" {
		return errors.New("access token is required")
	}
	if id.Currency == "" {
		return errors.New("currency is required")
	}
	return nil
}
