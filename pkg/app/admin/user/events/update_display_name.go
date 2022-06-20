package events

import (
	"context"
	"encoding/json"

	"github.com/grokloc/grokloc-server/pkg/safe"
)

type UpdateDisplayName struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

func (e *UpdateDisplayName) UnmarshalJSON(bs []byte) error {
	// clone type for a default unmarshal
	type updateDisplayNameEvent_ UpdateDisplayName
	var e_ updateDisplayNameEvent_
	err := json.Unmarshal(bs, &e_)
	if err != nil {
		return err
	}

	// use the fields from the default unmarshal to try to
	// construct a UpdateDisplayName
	n, err := NewUpdateDisplayName(
		context.Background(),
		e_.ID,
		e_.DisplayName,
	)
	if err != nil {
		return err
	}

	e.ID = n.ID
	e.DisplayName = n.DisplayName
	return nil
}

func NewUpdateDisplayName(
	ctx context.Context,
	id string,
	displayName string) (*UpdateDisplayName, error) {

	idErr := safe.IDIs(id)
	if idErr != nil {
		return nil, idErr
	}

	displayNameErr := safe.StringIs(displayName)
	if displayNameErr != nil {
		return nil, displayNameErr
	}

	return &UpdateDisplayName{
		ID:          id,
		DisplayName: displayName,
	}, nil
}
