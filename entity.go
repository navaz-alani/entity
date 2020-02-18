/*
Package entity defines a convenient abstraction which
can be used with MongoDB in order to streamline CRUD
operations.

This definition also allows a centralization of security
and other policies related to an entity within the app.
This leads to fewer bugs, and when there are bugs, due
to the modularization of policies, the origin of bugs
is easier to locate.

The goal is to define an abstraction which is useful for
the general entity and can make the process of writing code
much more efficient.

Axis Policy

An axis is defined as a field in an Entity which
can be assumed to be unique. This is important when creating
collection indexes and creating Query/Update/Delete filters.
The Axis Policy ensures data integrity by enforcing that all
Entities within a collection have unique axis values.


This Policy is especially useful when querying elements for
Read/Update/Delete operations. The client benefits through the
ability to specify whether a field is an axis, using the "axis"
tag, in the struct field. This tag can be set to "true" to enforce
it as an axis field.

Getting started

To use the Entity abstraction, start by creating a struct
which will define the Entity that you want to work with.
Available in this step, is the "axis" tag which is useful in
specifying which fields are to be treated as axis fields.
For example, here is a hypothetical struct for defining
a (useless) User Entity:

	type User struct {
		ID     primitive.ObjectID  `json:"-" bson:"_id"`
		Name   string              `json:"name" bson:"name"`
		Email  string              `json:"email" bson:"email" axis:"true" index:""`
	}

Next, register this User struct as an Entity.

	UserEntity := Entity{
		SchemaDefinition: TypeOf(User{}),
		PStorage:         &mongoCollection
	}

Run the Optimize function to generate indexes for the axis fields:

	UserEntity.Optimize()

Create a User:

	u := User{
		Name:  "Jane Doe",
		Email: "jane.doe@example.com"
	}

Add this user to the database:

	id, err := UserEntity.Add(u)

The other Read/Update/Delete operations become as simple with
this Entity definition and can lead to hundreds of lines of less
code being written.
The Entity package aims to (at least) provide a high level
API with the basic CRUD boilerplate code already taken care of.
*/
package entity

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/navaz-alani/entity/spec"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	JSONTag string = "json"
	BSONTag string = "bson"
	/*
		AxisTag is a tag used for tagging fields as
		axis fields for a particular struct.
	*/
	AxisTag string = "axis"
	/*
		IndexTag is used for tagging fields whose index
		needs to be created.
	*/
	IndexTag string = "index"
)

/*
EntityDefinition is a wrapper around the reflect.Type
and is used to define the schema for the Entities to be
collected in a collection.
*/
type EntityDefinition reflect.Type

/*
TypeOf returns an EntityDefinition which can be used with
an Entity to define a schema.
It performs a check to ensure that the entity is of kind
struct.
*/
func TypeOf(entity interface{}) EntityDefinition {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Struct {
		return entityType
	}
	return nil
}

/*
Filter uses the axis tags in a struct field to
create a BSON map which can be used to filter out
an entity from a collection.

The filter field is chosen with the following priority:
BSON tag "_id", Axis tag "true" (then BSON, JSON tags)
*/
func Filter(entity interface{}) bson.M {
	t := reflect.TypeOf(entity)
	v := reflect.ValueOf(entity)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		filterValue := v.Field(i).Interface()

		if tag := field.Tag.Get(BSONTag); tag == "_id" && filterValue != "" {
			return bson.M{"_id": filterValue}
		} else if tag := field.Tag.Get(AxisTag); tag == "true" && filterValue != "" {
			var filterFieldName string

			if tag := field.Tag.Get(BSONTag); tag != "" {
				filterFieldName = tag
			} else if tag := field.Tag.Get(JSONTag); tag != "" {
				filterFieldName = tag
			} else {
				filterFieldName = field.Name
			}

			return bson.M{filterFieldName: filterValue}
		}
	}

	return nil
}

/*
ToBSON returns a BSON map representing the given entity.
The given entity is expected to be of struct kind.

When converting, to BSON, field names are selected with
the following priority: BSON tag, JSON tag, field name
from the struct.
*/
func ToBSON(entity interface{}) bson.M {
	t := reflect.TypeOf(entity)
	v := reflect.ValueOf(entity)

	bsonEncoding := bson.M{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		var fieldName string
		if tag := field.Tag.Get(BSONTag); tag != "" {
			fieldName = tag
		} else if tag := field.Tag.Get(JSONTag); tag != "" {
			fieldName = tag
		} else {
			fieldName = field.Name
		}

		bsonEncoding[fieldName] = v.Field(i).Interface()
	}

	return bsonEncoding
}

/*
Entity is a type which is used to store
information about a collection of entities. It is
used to manage Entities and ensure persistence.
*/
type Entity struct {
	/*
		SchemaDefinition is the base type which will be
		used for this collection.
	*/
	SchemaDefinition EntityDefinition
	/*
		PStorage is the collection in which the Entities
		should be maintained.
	*/
	PStorage *mongo.Collection
}

/*
typeCheck verifies whether the entity can be used with the
Entity ec.
*/
func (e *Entity) typeCheck(entity interface{}) bool {
	if TypeOf(entity) != e.SchemaDefinition {
		return false
	}

	return true
}

/*
Add adds the given entity to the Entity ec.
The given entity is expected to be of struct kind.

This addition represents an actual insertion to the
underlying database collection pointed at by ec.

The added document's database ID is then returned, or
any errors that occurred.
*/
func (e *Entity) Add(entity interface{}) (primitive.ObjectID, error) {
	nilID := primitive.NilObjectID

	if !e.typeCheck(entity) {
		return nilID, fmt.Errorf("incompatible entity type")
	}

	dbDoc := ToBSON(entity)
	if dbDoc == nil {
		return nilID, fmt.Errorf(
			"entity body incomplete; will not add (Axis Policy)")
	}

	res, err := e.PStorage.InsertOne(context.TODO(), dbDoc)
	if err != nil {
		return nilID, err
	}

	addedID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nilID, fmt.Errorf("added entity but failed to parse addedID")
	}

	return addedID, nil
}

/*
Edit uses the axes of the given entity to find a
document in the underlying database collection pointed
at by ec and edits it according to the specified spec.

An error is returned which, if all went alright, should
be expected to be nil.
*/
func (e *Entity) Edit(entity interface{}, spec spec.ESpec) error {
	if !e.typeCheck(entity) {
		return fmt.Errorf("incompatible entity type")
	}

	filter := Filter(entity)
	if filter == nil {
		log.Panicln("entity axis undefined (Axis Policy)")
	}

	res := e.PStorage.FindOneAndUpdate(
		context.TODO(), filter, spec.ToUpdateSpec())
	return res.Err()
}

/*
Exists returns whether the filter produced by the given entity
matches any documents in the underlying database collection
pointed at by ec.
If any documents are matched and dest is non-nil, the matched
document will be decoded into dest, after which the fields can
be accessed.
If dest is left nil, the result is not decoded.

An error is also returned which, if all went alright, should
be expected to be nil.
*/
func (e *Entity) Exists(entity, dest interface{}) (bool, error) {
	if !e.typeCheck(entity) {
		return false, fmt.Errorf("incompatible entity type")
	}

	filter := Filter(entity)
	if filter == nil {
		return false, fmt.Errorf("entity axis undefined (Axis Policy)")
	}

	res := e.PStorage.FindOne(context.TODO(), filter)
	if res.Err() != mongo.ErrNoDocuments {
		if dest != nil {
			err := res.Decode(dest)
			if err != nil {
				return true, fmt.Errorf("failed to decode DB result")
			}
		}
		return true, nil
	}
	return false, res.Err()
}

/*
Delete deletes the given entity from the underlying database
collection pointed at by ec.

It returns an error from the delete operation which, if all
went well, can be expected to be nil.
*/
func (e *Entity) Delete(entity interface{}) error {
	if !e.typeCheck(entity) {
		return fmt.Errorf("incompatible entity type")
	}

	filter := Filter(entity)
	if filter == nil {
		return fmt.Errorf("entity axis undefined (Axis Policy)")
	}

	res := e.PStorage.FindOneAndDelete(context.TODO(), filter)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

/*
Optimize is a function that creates indexes for the axis fields
in the underlying EntityDefinition type.

Optimize searches for "index" tags in the fields of the type
underlying the EntityDefinition. A field with with an "index" tag
is optimized. The IndexModel entry for this field has the Key
corresponding to the BSON/JSON/field name (in that priority) and
value corresponding to the "index" tag value if non-empty and
a default index type of "text".
*/
func (e *Entity) Optimize() {
	keys := bson.D{}

	for i := 0; i < e.SchemaDefinition.NumField(); i++ {
		field := e.SchemaDefinition.Field(i)

		// Ignore field if IndexTag not set
		indexTag := field.Tag.Get(IndexTag)
		if indexTag == "" {
			continue
		}

		var fieldName string
		if tag := field.Tag.Get(BSONTag); tag != "" {
			fieldName = tag
		} else if tag := field.Tag.Get(JSONTag); tag != "" {
			fieldName = tag
		} else {
			fieldName = field.Name
		}

		var indexType string
		if indexTag != "" {
			indexType = indexTag
		} else {
			indexType = "text"
		}

		keys = append(keys, bson.E{fieldName, indexType})
	}

	index := []mongo.IndexModel{{Keys: keys}}

	opts := options.CreateIndexes().SetMaxTime(3 * time.Second)
	_, err := e.PStorage.Indexes().CreateMany(context.TODO(), index, opts)
	if err != nil {
		panic(err)
	}
}
