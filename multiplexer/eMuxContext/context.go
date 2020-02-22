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
	Payloads map[string]interface{}
}

/*
Create returns a pointer to an empty EMuxContext.
*/
func Create() *EMuxContext {
	payloadMap := make(map[string]interface{})
	return &EMuxContext{
		Payloads: payloadMap,
	}
}

/*
Set stores the given payload in the EMuxContext *emc
under the given key.
*/
func (emc *EMuxContext) Set(key string, payload interface{}) {
	emc.Payloads[key] = payload
}

/*
Get retrieves the payload stored under the given key
in the EMucContext *emc.
*/
func (emc *EMuxContext) Retrieve(key string) interface{} {
	return emc.Payloads[key]
}

/*
ContextualizeRequest returns the given request, with its context changed
to one with the given key and pointer to EMuxContext as the value.
*/
func (emc *EMuxContext) ContextualizeRequest(r *http.Request, parentCtx context.Context, key string) *http.Request {
	ctx := context.WithValue(parentCtx, key, emc)
	return r.WithContext(ctx)
}
