package admin

import (
	// "github.com/google/uuid"
	"github.com/grokloc/grokloc-server/pkg/models"
)

const OrgVersion = 0

type Org struct {
	models.Base
	Name  string `json:"name"`
	Owner string `json:"owner"`
}
