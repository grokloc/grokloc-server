package org

import "github.com/grokloc/grokloc-server/pkg/app"

type Controller struct {
	st *app.State
}

func NewController(st *app.State) (*Controller, error) {
	return &Controller{st: st}, nil
}
