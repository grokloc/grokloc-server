// Package org contains package methods for org support
package org

import "github.com/grokloc/grokloc-server/pkg/models"

type Org struct {
	models.Base
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

const Version = 0
