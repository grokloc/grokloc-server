package org

import (
	"context"
	"fmt"

	"github.com/grokloc/grokloc-server/pkg/grokloc"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
	"go.uber.org/zap"
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

func NewCreateEvent(
	ctx context.Context,
	name,
	ownerDisplayName,
	ownerEmail,
	ownerPassword string) (*CreateEvent, error) {

	defer func() {
		_ = zap.L().Sync()
	}()

	args := map[string]string{
		"owner display name": ownerDisplayName,
		"owner email":        ownerEmail,
		"owner password":     ownerPassword,
	}

	for k, v := range args {
		if !security.SafeStr(v) {
			zap.L().Info(fmt.Sprintf("%s unsafe", k),
				zap.Error(models.ErrUnsafeString),
				zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
			)
			return nil, models.ErrUnsafeString
		}
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

	defer func() {
		_ = zap.L().Sync()
	}()

	args := map[string]string{
		"id":    id,
		"owner": owner,
	}

	for k, v := range args {
		if !security.SafeStr(v) {
			zap.L().Info(fmt.Sprintf("%s unsafe", k),
				zap.Error(models.ErrUnsafeString),
				zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
			)
			return nil, models.ErrUnsafeString
		}
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

	defer func() {
		_ = zap.L().Sync()
	}()

	args := map[string]string{
		"id": id,
	}

	for k, v := range args {
		if !security.SafeStr(v) {
			zap.L().Info(fmt.Sprintf("%s unsafe", k),
				zap.Error(models.ErrUnsafeString),
				zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
			)
			return nil, models.ErrUnsafeString
		}
	}

	status, err := models.NewStatus(statusInt)
	if err != nil {
		zap.L().Info(fmt.Sprintf("%v not acceptable status", statusInt),
			zap.Error(models.ErrDisallowedValue),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, models.ErrDisallowedValue
	}

	if status == models.StatusUnconfirmed {
		zap.L().Info("cannot set existing row to unconfirmed",
			zap.Error(models.ErrStatus),
			zap.String(grokloc.RequestIDKey, grokloc.CtxRequestID(ctx)),
		)
		return nil, models.ErrStatus
	}

	return &UpdateStatusEvent{
		ID:     id,
		Status: status,
	}, nil
}
