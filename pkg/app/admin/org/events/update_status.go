package events

import (
	"context"
	"encoding/json"

	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/safe"
)

type UpdateStatus struct {
	ID     string        `json:"id"`
	Status models.Status `json:"status"`
}

func (e *UpdateStatus) UnmarshalJSON(bs []byte) error {
	// clone type for a default unmarshal
	type updateStatusEvent_ UpdateStatus
	var e_ updateStatusEvent_
	err := json.Unmarshal(bs, &e_)
	if err != nil {
		return err
	}

	// use the fields from the default unmarshal to try to
	// construct a UpdateStatus
	n, err := NewUpdateStatus(
		context.Background(),
		e_.ID,
		int(e_.Status),
	)
	if err != nil {
		return err
	}

	// all fields are safe, assign to use
	e.ID = n.ID
	e.Status = n.Status
	return nil
}

func NewUpdateStatus(
	ctx context.Context,
	id string,
	statusInt int) (*UpdateStatus, error) {

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

	return &UpdateStatus{
		ID:     id,
		Status: status,
	}, nil
}
