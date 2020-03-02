/*
Package muxContext defines a simple concurrency-safe context that
can be used with HTTP requests to easily store multiple pieces of
information within the same http.Request context.
*/
package muxContext

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/navaz-alani/entity/entityErrors"
)

/*
muxCtxKey is the keyStr which maps to EMux stored information within
an http.Request with a context.
*/
const muxCtxKey = "_muxCtx_"

/*
EMuxContext is a simple map used to organize multiple pieces
of information within one http.Request context.
It has been written to be concurrency safe under read/write
operations using the methods provided.
*/
type EMuxContext struct {
	/*
		payloads is internally used to map keys to payloads.
	*/
	payloads map[string]interface{}
	/*
		err is used to store information about the results of
		the pre-processing stage.
	*/
	err error
	/*
		mutex is used to internally ensure that concurrent
		read/write operations do not compromise payload data.
	*/
	mutex *sync.Mutex
}

/*
Create initializes and returns a pointer to an new EMuxContext.
*/
func Create() *EMuxContext {
	payloadMap := make(map[string]interface{})

	return &EMuxContext{
		payloads: payloadMap,
		mutex:    &sync.Mutex{},
	}
}

/*
Set stores the given payload in the EMuxContext *emc
under the given keyStr.
*/
func (emc *EMuxContext) Set(key string, payload interface{}) error {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	emc.payloads[key] = payload
	return nil
}

/*
Get retrieves the payload stored under the given keyStr
in the EMucContext *emc.
*/
func (emc *EMuxContext) Retrieve(key string) interface{} {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	return emc.payloads[key]
}

/*
SetError sets the error state of emc.
*/
func (emc *EMuxContext) SetError(msg string) {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	emc.err = fmt.Errorf(msg)
}

/*
Error retrieves the error state of emc.
*/
func (emc *EMuxContext) Error() error {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	return emc.err
}

/*
Embed returns the given request, with its context modified
to include emc. Internally, a pointer to the context is
stored so emc can be changed after embedding it into the
request.
*/
func (emc *EMuxContext) Embed(r *http.Request, parentCtx context.Context) *http.Request {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	ctx := context.WithValue(parentCtx, muxCtxKey, emc)
	return r.WithContext(ctx)
}

/*
Isolate returns a pointer to the EMuxContext which is stored
within the context of the given request, and any errors
associated with the operation.

If the context does not exist, entityErrors.MuxCtxNotFound is
returned.
If the context cannot be parsed, entityErrors.MuxCtxCorrupt is
returned.
*/
func Isolate(r *http.Request) (*EMuxContext, error) {
	reqCtxVal := r.Context().Value(muxCtxKey)
	if reqCtxVal == nil {
		return nil, entityErrors.MuxCtxNotFound
	}

	if muxCtx, ok := reqCtxVal.(*EMuxContext); !ok {
		return nil, entityErrors.MuxCtxCorrupt
	} else {
		return muxCtx, nil
	}
}
