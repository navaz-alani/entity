/*
Package muxHandle defines an interface which specifies
database behaviour required by the multiplexer.
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

This definition helps make the multiplexer testable,
as well as constrain the multiplexer's capabilities
towards the underlying database.
*/
type DBHandler interface {
	/*
		Collection is a method required by the multiplexer
		to obtain handles for persistent storage for Entities.
	*/
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
}

type CollectionHandler interface {
	InsertOne(ctx context.Context, document interface{},
		opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOneAndDelete(ctx context.Context, filter interface{},
		opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult
	FindOneAndUpdate(ctx context.Context, filter interface{},
		update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
	FindOne(ctx context.Context, filter interface{},
		opts ...*options.FindOneOptions) *mongo.SingleResult
	Indexes() mongo.IndexView
}

type Creator interface {
	InsertOne(ctx context.Context, document interface{},
		opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
}

type Deleter interface {
	FindOneAndDelete(ctx context.Context, filter interface{},
		opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult
}

type Editor interface {
	FindOneAndUpdate(ctx context.Context, filter interface{},
		update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
}

type Finder interface {
	FindOne(ctx context.Context, filter interface{},
		opts ...*options.FindOneOptions) *mongo.SingleResult
}

type Indexer interface {
	Indexes() mongo.IndexView
}
