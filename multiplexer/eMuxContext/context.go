/*
Package eMuxContext defines a simple context that can be used
with HTTP requests to easily store multiple pieces of information
within the same http.Request context.
*/
package eMuxContext

import (
	"context"
	"net/http"
)

/*
EMuxKey is the key which maps to EMux stored information within
an http.Request with a context.
*/
const EMuxKey = "eMux"

/*
EMuxContext is a simple map used to organize multiple pieces
of information within one http.Request context.
*/
type EMuxContext struct {
	Payloads *map[string]interface{}
}

/*
PackagePayload sets the value of the emc.Payloads entry corresponding to
the given key to contain the given payload
*/
func (emc *EMuxContext) PackagePayload(key string, payload interface{}) {
	(*emc.Payloads)[key] = payload
}

/*
ContextualizedRequest returns the given request, with its context changed
to one with the given key and pointer to EMuxContext as the value.
*/
func (emc *EMuxContext) ContextualizedRequest(r *http.Request, parentCtx context.Context, key string) *http.Request {
	ctx := context.WithValue(parentCtx, key, emc)
	return r.WithContext(ctx)
}
