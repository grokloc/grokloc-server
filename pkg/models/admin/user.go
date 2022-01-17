package admin

import (
	// "github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/models"
)

const UserVersion = 0

type User struct {
	models.Base
	APISecret         string `json:"api_secret"`
	APISecretDigest   string `json:"api_secret_digest"`
	DisplayName       string `json:"display_name"`
	DisplayNameDigest string `json:"display_name_digest"`
	Email             string `json:"email"`
	EmailDigest       string `json:"email_digest"`
	Org               string `json:"org"`
	Password          string `json:"-"`
}
