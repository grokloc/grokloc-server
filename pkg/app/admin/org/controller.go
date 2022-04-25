package org

import (
	"context"

	"github.com/grokloc/grokloc-server/pkg/app"
	"github.com/grokloc/grokloc-server/pkg/grokloc"
	"go.uber.org/zap"
)

type Controller struct {
	state *app.State
}

func NewController(ctx context.Context, state *app.State) (*Controller, error) {
	return &Controller{state: state}, nil
}

func (c *Controller) Create(ctx context.Context, event CreateEvent) (*Org, error) {

	defer func() {
		_ = zap.L().Sync()
	}()

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
		zap.L().Error("org::Controller::Create",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, err
	}

	return org, nil
}

func (c *Controller) Read(ctx context.Context, id string) (*Org, error) {

	defer func() {
		_ = zap.L().Sync()
	}()

	return Read(ctx, id, c.state.RandomReplica())
}

func (c *Controller) UpdateOwner(ctx context.Context, event UpdateOwnerEvent) (*Org, error) {

	defer func() {
		_ = zap.L().Sync()
	}()

	org, err := c.Read(ctx, event.ID)

	if err != nil {
		zap.L().Error("org::Controller::UpdateOwner",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, err
	}

	err = org.UpdateOwner(ctx, event.Owner, c.state.Master)
	if err != nil {
		zap.L().Error("org::Controller::UpdateOwner",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, err
	}

	return org, nil
}

func (c *Controller) UpdateStatus(ctx context.Context, event UpdateStatusEvent) (*Org, error) {

	defer func() {
		_ = zap.L().Sync()
	}()

	org, err := c.Read(ctx, event.ID)

	if err != nil {
		zap.L().Error("org::Controller::UpdateStatus",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, err
	}

	err = org.UpdateStatus(ctx, event.Status, c.state.Master)
	if err != nil {
		zap.L().Error("org::Controller::UpdateStatus",
			zap.Error(err),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, err
	}

	return org, nil
}
