package events

import (
	"context"

	org_event "github.com/grokloc/grokloc-server/pkg/app/admin/org/events"
)

// UpdateStatus is identical to, and operates identically to, the org event
type UpdateStatus org_event.UpdateStatus

func NewUpdateStatus(
	ctx context.Context,
	id string,
	statusInt int) (*UpdateStatus, error) {

	event, err := org_event.NewUpdateStatus(ctx, id, statusInt)
	if err != nil {
		return nil, err
	}

	return &UpdateStatus{ID: id, Status: event.Status}, nil
}
