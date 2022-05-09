package events

import (
	"context"
	"encoding/json"

	"github.com/grokloc/grokloc-server/pkg/safe"
)

type Create struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Org         string `json:"org"`
	// Password assumed already derived
	Password string `json:"password"`
}

func (e *Create) UnmarshalJSON(bs []byte) error {
	// clone type for a default unmarshal
	type createEvent_ Create
	var e_ createEvent_
	err := json.Unmarshal(bs, &e_)
	if err != nil {
		return err
	}

	// use the fields from the default unmarshal to try to
	// construct a Create
	n, err := NewCreate(
		context.Background(),
		e_.DisplayName,
		e_.Email,
		e_.Org,
		e_.Password,
	)
	if err != nil {
		return err
	}

	e.DisplayName = n.DisplayName
	e.Email = n.Email
	e.Org = n.Org
	e.Password = n.Password
	return nil
}

func NewCreate(
	ctx context.Context,
	displayName,
	email,
	org,
	password string) (*Create, error) {

	displayNameErr := safe.StringIs(displayName)
	if displayNameErr != nil {
		return nil, displayNameErr
	}

	emailErr := safe.StringIs(email)
	if emailErr != nil {
		return nil, emailErr
	}

	orgErr := safe.StringIs(org)
	if orgErr != nil {
		return nil, orgErr
	}

	passwordErr := safe.StringIs(password)
	if passwordErr != nil {
		return nil, passwordErr
	}

	return &Create{
		DisplayName: displayName,
		Email:       email,
		Org:         org,
		Password:    password,
	}, nil
}
