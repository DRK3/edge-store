/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package edv

import (
	"github.com/trustbloc/edge-store/pkg/restapi/edv/operation"
	"github.com/trustbloc/edge-store/pkg/storage"
)

// New returns new controller instance.
func New(provider storage.Provider) (*Controller, error) {
	var allHandlers []operation.Handler

	edvService := operation.New(provider)
	allHandlers = append(allHandlers, edvService.GetRESTHandlers()...)

	return &Controller{handlers: allHandlers}, nil
}

// Controller contains handlers for controller
type Controller struct {
	handlers []operation.Handler
}

// GetOperations returns all controller endpoints
func (c *Controller) GetOperations() []operation.Handler {
	return c.handlers
}
