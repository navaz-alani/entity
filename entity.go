package entity

import (
	"context"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/navaz-alani/entity/eField"
	"github.com/navaz-alani/entity/entityErrors"
	"github.com/navaz-alani/entity/spec"
)

const (
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
	/*
		IDTag is used for providing Entity identifiers
		for an EntityMux.
	*/
	IDTag string = "_id_"
	/*
		HandleTag is used to provide configuration for
		generating pre-processing middleware.
	*/
	HandleTag string = "_hd_"
)

/*
TypeOf returns an EntityDefinition which can be used with
an Entity to define a schema.
It performs a check to ensure that the entity is of kind
struct.
*/
func TypeOf(entity interface{}) reflect.Type {
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() == reflect.Struct {
		return entityType
	}
	return nil
}

/*
Filter uses the axis tags in a struct eField to
create a BSON map which can be used to filter out
an entity from a collection.

The filter eField is chosen with the following priority:
BSON tag "_id", Axis tag "true" (then BSON, JSON tags)
and lastly the eField name.

Note that the eField with the BSON tag "_id" must be of
type primitive.ObjectID so that comparison succeeds.
*/
func Filter(entity interface{}) bson.M {
	t := reflect.TypeOf(entity)
	v := reflect.ValueOf(entity)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		filterValue := v.Field(i).Interface()

		if tag := field.Tag.Get(BSONTag); tag == "_id" && filterValue != primitive.NilObjectID {
			return bson.M{"_id": filterValue}
		} else if tag := field.Tag.Get(AxisTag); tag == "true" && filterValue != "" {
			var filterFieldName = eField.NameByPriority(field, eField.PriorityBsonJson)
			return bson.M{filterFieldName: filterValue}
		}
	}

	return nil
}

/*
ToBSON returns a BSON map representing the given entity.
The given entity is expected to be of struct kind.

When converting, to BSON, eField names are selected with
the following priority: BSON tag, JSON tag, eField name
from the struct.
*/
func ToBSON(entity interface{}) bson.M {
	t := reflect.TypeOf(entity)
	v := reflect.ValueOf(entity)

	bsonEncoding := bson.M{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get(BSONTag); tag == "_id" {
			continue
		}

		var fName = eField.NameByPriority(field, eField.PriorityBsonJson)

		bsonEncoding[fName] = v.Field(i).Interface()
	}

	return bsonEncoding
}

/*
Entity is a type which is used to store
information about a collection of entities. It is
used to manage Entities and ensure persistence.

The SchemaDefinition eField's contents is used to
generate a validator for the collection. This is
done using "validate" tags which allow deeper
schema specification.
*/
type Entity struct {
	/*
		SchemaDefinition is the base type which will be
		used for this collection.
	*/
	SchemaDefinition reflect.Type
	/*
		PStorage is the collection in which the Entities
		should be maintained.
	*/
	PStorage *mongo.Collection
}

/*
typeCheck verifies whether the entity can be used with the
Entity e.
*/
func (e *Entity) typeCheck(entity interface{}) bool {
	return TypeOf(entity) == e.SchemaDefinition
}

/*
Add adds the given entity to the Entity e.
The given entity is expected to be of struct kind.

This addition represents an actual insertion to the
underlying database collection pointed at by e.

The added document's database ID is then returned, or
any entityErrors that occurred.
*/
func (e *Entity) Add(entity interface{}) (primitive.ObjectID, error) {
	nilID := primitive.NilObjectID

	if !e.typeCheck(entity) {
		return nilID, entityErrors.IncompatibleEntityType
	}

	dbDoc := ToBSON(entity)
	if dbDoc == nil || len(dbDoc) == 0 {
		return nilID, entityErrors.BodyIncomplete
	}

	// TODO: add check for whether the defined axis fields are unique

	res, err := e.PStorage.InsertOne(context.TODO(), dbDoc)
	if err != nil {
		return nilID, err
	}

	addedID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nilID, entityErrors.AddedIDParseFail
	}

	return addedID, nil
}

/*
Edit uses the axes of the given entity to find a
document in the underlying database collection pointed
at by e and edits it according to the specified spec.

An error is returned which, if all went alright, should
be expected to be nil.
*/
func (e *Entity) Edit(entity interface{}, spec spec.ESpec) error {
	if !e.typeCheck(entity) {
		return entityErrors.IncompatibleEntityType
	}

	filter := Filter(entity)
	if filter == nil {
		return entityErrors.UndefinedAxis
	}

	res := e.PStorage.FindOneAndUpdate(
		context.TODO(), filter, spec.ToUpdateSpec())
	return res.Err()
}

/*
Exists returns whether the filter produced by the given entity
matches any documents in the underlying database collection
pointed at by e.
If any documents are matched and dest is non-nil, the matched
document will be decoded into dest, after which the fields can
be accessed.
If dest is left nil, the result is not decoded.

An error is also returned which, if all went alright, should
be expected to be nil.
*/
func (e *Entity) Exists(entity, dest interface{}) (bool, error) {
	if !e.typeCheck(entity) {
		return false, entityErrors.IncompatibleEntityType
	}

	filter := Filter(entity)
	if filter == nil {
		return false, entityErrors.UndefinedAxis
	}

	res := e.PStorage.FindOne(context.TODO(), filter)
	if res.Err() != mongo.ErrNoDocuments {
		if dest != nil {
			err := res.Decode(dest)
			if err != nil {
				return true, entityErrors.DBDecodeFail
			}

			return true, nil
		}
		return true, nil
	}
	return false, res.Err()
}

/*
Delete deletes the given entity from the underlying database
collection pointed at by e.

It returns an error from the delete operation which, if all
went well, can be expected to be nil.
*/
func (e *Entity) Delete(entity interface{}) error {
	if !e.typeCheck(entity) {
		return entityErrors.IncompatibleEntityType
	}

	filter := Filter(entity)
	if filter == nil {
		return entityErrors.UndefinedAxis
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
underlying the EntityDefinition. A eField with with an "index" tag
is optimized. The IndexModel entry for this eField has the Key
corresponding to the BSON/JSON/eField name (in that priority) and
value corresponding to the "index" tag value if non-empty and
a default index type of "text".
*/
func (e *Entity) Optimize() error {
	keys := bson.D{}

	for i := 0; i < e.SchemaDefinition.NumField(); i++ {
		field := e.SchemaDefinition.Field(i)

		// Ignore eField if IndexTag not set
		indexTag := field.Tag.Get(IndexTag)
		axisTag := field.Tag.Get(AxisTag)
		if !(indexTag == "true" && axisTag == "true") {
			continue
		}

		var key = eField.NameByPriority(field, eField.PriorityBsonJson)

		var indexType string
		if !(indexTag == "" || indexTag == "-") {
			indexType = indexTag
		} else {
			// TODO: infer index type from eField type
			indexType = "text"
		}

		keys = append(keys, bson.E{Key: key, Value: indexType})
	}

	if len(keys) == 0 {
		return nil
	}
	index := []mongo.IndexModel{{Keys: keys}}

	opts := options.CreateIndexes().SetMaxTime(3 * time.Second)
	_, err := e.PStorage.Indexes().CreateMany(context.TODO(), index, opts)
	if err != nil {
		return err
	}
	return nil
}
