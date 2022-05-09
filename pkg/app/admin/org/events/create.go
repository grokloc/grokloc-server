package events

import (
	"context"
	"encoding/json"

	"github.com/grokloc/grokloc-server/pkg/safe"
)

type Create struct {
	Name             string `json:"name"`
	OwnerDisplayName string `json:"owner_display_name"`
	OwnerEmail       string `json:"owner_email"`
	// OwnerPassword assumed already derived
	OwnerPassword string `json:"owner_password"`
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
		e_.Name,
		e_.OwnerDisplayName,
		e_.OwnerEmail,
		e_.OwnerPassword,
	)
	if err != nil {
		return err
	}

	e.Name = n.Name
	e.OwnerDisplayName = n.OwnerDisplayName
	e.OwnerEmail = n.OwnerEmail
	e.OwnerPassword = n.OwnerPassword
	return nil
}

func NewCreate(
	ctx context.Context,
	name,
	ownerDisplayName,
	ownerEmail,
	ownerPassword string) (*Create, error) {

	nameErr := safe.StringIs(name)
	if nameErr != nil {
		return nil, nameErr
	}

	ownerDisplayNameErr := safe.StringIs(ownerDisplayName)
	if ownerDisplayNameErr != nil {
		return nil, ownerDisplayNameErr
	}

	ownerEmailErr := safe.StringIs(ownerEmail)
	if ownerEmailErr != nil {
		return nil, ownerEmailErr
	}

	ownerPasswordErr := safe.StringIs(ownerPassword)
	if ownerPasswordErr != nil {
		return nil, ownerPasswordErr
	}

	return &Create{
		Name:             name,
		OwnerDisplayName: ownerDisplayName,
		OwnerEmail:       ownerEmail,
		OwnerPassword:    ownerPassword,
	}, nil
}
