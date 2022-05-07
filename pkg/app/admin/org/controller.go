package org

import (
	"context"

	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/app/admin/org/events"
)

type Controller struct {
	state *app.State
}

func NewController(ctx context.Context, state *app.State) (*Controller, error) {
	return &Controller{state: state}, nil
}

func (c *Controller) Create(ctx context.Context, event events.Create) (*Org, error) {

	org, err := Create(
		ctx,
		event.Name,
		event.OwnerDisplayName,
		event.OwnerEmail,
		event.OwnerPassword,
		c.state.DBKey,
		c.state.Master,
	)

	if err != nil {
		return nil, err
	}

	return org, nil
}

func (c *Controller) Read(ctx context.Context, id string) (*Org, error) {
	return Read(ctx, id, c.state.RandomReplica())
}

func (c *Controller) UpdateOwner(ctx context.Context, event events.UpdateOwner) (*Org, error) {

	org, err := c.Read(ctx, event.ID)

	if err != nil {
		return nil, err
	}

	err = org.UpdateOwner(ctx, event.Owner, c.state.Master)
	if err != nil {
		return nil, err
	}

	return org, nil
}

func (c *Controller) UpdateStatus(ctx context.Context, event events.UpdateStatus) (*Org, error) {

	org, err := c.Read(ctx, event.ID)

	if err != nil {
		return nil, err
	}

	err = org.UpdateStatus(ctx, event.Status, c.state.Master)
	if err != nil {
		return nil, err
	}

	return org, nil
}
