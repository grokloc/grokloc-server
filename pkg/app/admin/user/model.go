// Package user contains package methods for user support
package user

import "github.com/grokloc/grokloc-server/pkg/models"

type User struct {
	models.Base
	APISecret         string `json:"api_secret"`
	APISecretDigest   string `json:"api_secret_digest"`
	DisplayName       string `json:"display_name"`
	DisplayNameDigest string `json:"display_name_digest"`
	Email             string `json:"email"`
	EmailDigest       string `json:"email_digest"`
	Org               string `json:"org"`
	// Password is assumed initialized as derived
	Password string `json:"-"`
}

const Version = 0
