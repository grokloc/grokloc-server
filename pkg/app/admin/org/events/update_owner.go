package events

import (
	"context"
	"encoding/json"

	"github.com/grokloc/grokloc-server/pkg/safe"
)

type UpdateOwner struct {
	ID    string `json:"id"`
	Owner string `json:"owner"`
}

func (e *UpdateOwner) UnmarshalJSON(bs []byte) error {
	// clone type for a default unmarshal
	type updateOwnerEvent_ UpdateOwner
	var e_ updateOwnerEvent_
	err := json.Unmarshal(bs, &e_)
	if err != nil {
		return err
	}

	// use the fields from the default unmarshal to try to
	// construct a UpdateOwner
	n, err := NewUpdateOwner(
		context.Background(),
		e_.ID,
		e_.Owner,
	)
	if err != nil {
		return err
	}

	e.ID = n.ID
	e.Owner = n.Owner
	return nil
}

func NewUpdateOwner(
	ctx context.Context,
	id string,
	owner string) (*UpdateOwner, error) {

	idErr := safe.StringIs(id)
	if idErr != nil {
		return nil, idErr
	}

	ownerErr := safe.StringIs(owner)
	if ownerErr != nil {
		return nil, ownerErr
	}

	return &UpdateOwner{
		ID:    id,
		Owner: owner,
	}, nil
}
