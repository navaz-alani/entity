package entityErrors

import "fmt"

var pkgPath = "entity/multiplexer/muxContext"

var MuxCtxNotFound = fmt.Errorf("%s: EMuxContext not found in request", pkgPath)
var MuxCtxCorrupt = fmt.Errorf("%s: reteived EMuxContext corrupt", pkgPath)