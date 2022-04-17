package org

type CreateEvent struct {
	Name             string `json:"name"`
	OwnerDisplayName string `json:"owner_display_name"`
	OwnerEmail       string `json:"owner_email"`
	// OwnerPassword assumed already derived
	OwnerPassword string `json:"owner_password"`
}
