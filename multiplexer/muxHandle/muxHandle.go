/*
Package muxHandle defines an interface which specifies
database behaviour required by the multiplexer.

These interfaces abstract the database layer out so
that package functions can be tested.
*/
package muxHandle

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
DBHandler is an interface which defines the behaviour
of a database handle required by the multiplexer.
*/
type DBHandler interface {
	/*
		Collection is a method required by the multiplexer
		to obtain handles for persistent storage for Entities.
	*/
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
}

/*
EntityCollector is an interface describing the functionality
required by the Entity for operations on specific collections.
*/
type EntityCollector interface {
	// Creation
	InsertOne(ctx context.Context, document interface{},
		opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	// Reading
	FindOne(ctx context.Context, filter interface{},
		opts ...*options.FindOneOptions) *mongo.SingleResult
	// Updating
	FindOneAndUpdate(ctx context.Context, filter interface{},
		update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
	// Deleting
	FindOneAndDelete(ctx context.Context, filter interface{},
		opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult

	// Indexing
	Indexes() mongo.IndexView
}
