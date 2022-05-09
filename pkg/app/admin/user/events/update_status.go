package events

import (
	org_event "github.com/grokloc/grokloc-server/pkg/app/admin/org/events"
)

// UpdateStatus is identical to, and operates identically to, the org event
type UpdateStatus org_event.UpdateStatus
