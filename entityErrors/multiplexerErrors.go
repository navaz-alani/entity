/*
Package entityErrors provides a set of default errors for
repeated use within the library.
*/
package entityErrors

import "fmt"

var DBUninitialized = fmt.Errorf("null db pointer")
var IncompleteEntityMetadata = fmt.Errorf("insufficient entity metadata")

func NoTag(tag, entity string) error {
	return fmt.Errorf("no '%s' tag on '%s'", tag, entity)
}

func DuplicateTag(tag, entity string) error {
	return fmt.Errorf("duplicate '%s' tag on '%s'", tag, entity)
}
