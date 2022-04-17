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

func NewController(state *app.State) (*Controller, error) {
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
