/*
Package muxHandle defines an interface which specifies
database behaviour required by the multiplexer.
*/
package muxHandle

import (
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
DBHandle is an interface which defines the behaviour
of a database handle required by the multiplexer.

This definition helps make the multiplexer testable,
as well as constrain the multiplexer's operations
on the database.
*/
type DBHandle interface {
	/*
		Collection is a method required by the multiplexer
		to obtain handles for persistent storage for Entities.
	*/
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
}
