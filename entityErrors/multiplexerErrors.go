/*
Package entityErrors provides a set of default errors for
repeated use within the library.
*/
package entityErrors

import "fmt"

var (
	/*
		DBUninitialized is an error returned when a database
		pointer is nil.
	*/
	DBUninitialized = fmt.Errorf("null db pointer")
	/*
		IncompleteEntityMetadata is an error which signifies
		that an Entity's profile is incomplete. This could be
		as a result of undefined tags or tags with empty values.
	*/
	IncompleteEntityMetadata = fmt.Errorf("insufficient entity metadata")
	NoClassificationFields   = fmt.Errorf("no classification fields")
	InvalidDataType = fmt.Errorf("data type invalid invalid")
)

/*
NoTag is an error representing the absence of a required
tag for a particular operation.
*/
func NoTag(tag, entity string) error {
	return fmt.Errorf("no '%s' tag on '%s'", tag, entity)
}

/*
DuplicateTag is an error representing that a tag, which
needs to have distinct values across Entities, has been
found to have duplicates.
*/
func DuplicateTag(tag, entity string) error {
	return fmt.Errorf("duplicate '%s' tag on '%s'", tag, entity)
}
