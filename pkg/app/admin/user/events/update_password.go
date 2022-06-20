package events

import (
	"context"
	"encoding/json"

	"github.com/grokloc/grokloc-server/pkg/safe"
)

type UpdatePassword struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

func (e *UpdatePassword) UnmarshalJSON(bs []byte) error {
	// clone type for a default unmarshal
	type updatePasswordEvent_ UpdatePassword
	var e_ updatePasswordEvent_
	err := json.Unmarshal(bs, &e_)
	if err != nil {
		return err
	}

	// use the fields from the default unmarshal to try to
	// construct a UpdatePassword
	n, err := NewUpdatePassword(
		context.Background(),
		e_.ID,
		e_.Password,
	)
	if err != nil {
		return err
	}

	e.ID = n.ID
	e.Password = n.Password
	return nil
}

func NewUpdatePassword(
	ctx context.Context,
	id string,
	password string) (*UpdatePassword, error) {

	idErr := safe.IDIs(id)
	if idErr != nil {
		return nil, idErr
	}

	passwordErr := safe.StringIs(password)
	if passwordErr != nil {
		return nil, passwordErr
	}

	return &UpdatePassword{
		ID:       id,
		Password: password,
	}, nil
}
