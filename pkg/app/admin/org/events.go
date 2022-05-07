package org

import (
	"context"
	"encoding/json"

	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/safe"
)

// events are defined for mutations, with constructors to validate
// new states

type CreateEvent struct {
	Name             string `json:"name"`
	OwnerDisplayName string `json:"owner_display_name"`
	OwnerEmail       string `json:"owner_email"`
	// OwnerPassword assumed already derived
	OwnerPassword string `json:"owner_password"`
}

func (ce *CreateEvent) UnmarshalJSON(bs []byte) error {
	// clone type CreateEvent for a default unmarshal
	type createEvent_ CreateEvent
	var ce_ createEvent_
	err := json.Unmarshal(bs, &ce_)
	if err != nil {
		return err
	}

	// use the fields from the default unmarshal to try to
	// construct a CreateEvent
	nce, err := NewCreateEvent(
		context.Background(),
		ce_.Name,
		ce_.OwnerDisplayName,
		ce_.OwnerEmail,
		ce_.OwnerPassword,
	)
	if err != nil {
		return nil
	}

	// all fields are safe, assign to ce
	ce.Name = nce.Name
	ce.OwnerDisplayName = nce.OwnerDisplayName
	ce.OwnerEmail = nce.OwnerEmail
	ce.OwnerPassword = nce.OwnerPassword
	return nil
}

func NewCreateEvent(
	ctx context.Context,
	name,
	ownerDisplayName,
	ownerEmail,
	ownerPassword string) (*CreateEvent, error) {

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

	return &CreateEvent{
		Name:             name,
		OwnerDisplayName: ownerDisplayName,
		OwnerEmail:       ownerEmail,
		OwnerPassword:    ownerPassword,
	}, nil
}

type UpdateOwnerEvent struct {
	ID    string `json:"id"`
	Owner string `json:"owner"`
}

func NewUpdateOwner(
	ctx context.Context,
	id string,
	owner string,
	ownerPassword string) (*UpdateOwnerEvent, error) {

	idErr := safe.StringIs(id)
	if idErr != nil {
		return nil, idErr
	}

	ownerErr := safe.StringIs(owner)
	if ownerErr != nil {
		return nil, ownerErr
	}

	return &UpdateOwnerEvent{
		ID:    id,
		Owner: owner,
	}, nil
}

type UpdateStatusEvent struct {
	ID     string        `json:"id"`
	Status models.Status `json:"status"`
}

func NewUpdateStatusEvent(
	ctx context.Context,
	id string,
	statusInt int) (*UpdateStatusEvent, error) {

	idErr := safe.StringIs(id)
	if idErr != nil {
		return nil, idErr
	}

	status, err := models.NewStatus(statusInt)
	if err != nil {
		return nil, models.ErrDisallowedValue
	}

	if status == models.StatusUnconfirmed {
		return nil, models.ErrStatus
	}

	return &UpdateStatusEvent{
		ID:     id,
		Status: status,
	}, nil
}
