package user

import (
	"context"

	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user/events"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
)

type Controller struct {
	state *app.State
}

func NewController(ctx context.Context, state *app.State) (*Controller, error) {
	return &Controller{state: state}, nil
}

func (c *Controller) Create(ctx context.Context, event events.Create) (*User, error) {

	// password asumed clear text - derive it
	password, err := security.DerivePassword(
		event.Password,
		c.state.Argon2Cfg,
	)
	if err != nil {
		return nil, err
	}

	user, err := Create(
		ctx,
		event.DisplayName,
		event.Email,
		event.Org,
		password,
		c.state.DBKey,
		c.state.Master,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *Controller) Read(ctx context.Context, id string) (*User, error) {
	return Read(ctx, id, c.state.DBKey, c.state.RandomReplica())
}

func (c *Controller) UpdateDisplayName(ctx context.Context, event events.UpdateDisplayName) (*User, error) {

	user, err := c.Read(ctx, event.ID)

	if err != nil {
		return nil, err
	}

	// no user found with ID
	if user == nil {
		return nil, models.ErrNotFound
	}

	err = user.UpdateDisplayName(ctx, event.DisplayName, c.state.DBKey, c.state.Master)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *Controller) UpdatePassword(ctx context.Context, event events.UpdatePassword) (*User, error) {

	user, err := c.Read(ctx, event.ID)

	if err != nil {
		return nil, err
	}

	// no user found with ID
	if user == nil {
		return nil, models.ErrNotFound
	}

	// password asumed clear text - derive it
	password, err := security.DerivePassword(
		event.Password,
		c.state.Argon2Cfg,
	)
	if err != nil {
		return nil, err
	}

	err = user.UpdatePassword(ctx, password, c.state.Master)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *Controller) UpdateStatus(ctx context.Context, event events.UpdateStatus) (*User, error) {

	user, err := c.Read(ctx, event.ID)

	if err != nil {
		return nil, err
	}

	// no user found with ID
	if user == nil {
		return nil, models.ErrNotFound
	}

	err = user.UpdateStatus(ctx, event.Status, c.state.Master)
	if err != nil {
		return nil, err
	}

	return user, nil
}
