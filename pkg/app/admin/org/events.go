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
