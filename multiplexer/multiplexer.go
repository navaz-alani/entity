/*
Package multiplexer defines an EntityMux type which is basically
a multiplexer for Entity types.
*/
package multiplexer

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/navaz-alani/entity"
)

/*
EntityMux is a multiplexer for Entities.
*/
type EntityMux struct {
	/*
		entities is a collection of Entities which are
		used in an application.
		In the entities map, the key is the Entity's ID
		field value.
	*/
	entities map[string]*entity.Entity
}

/*
Collection returns a pointer to the mongo collection
that the entity identified by the given entityID is
using.
*/
func (em *EntityMux) Collection(entityID string) *mongo.Collection {
	return em.entities[entityID].PStorage
}

/*
Create uses the given database and data definitions
to initialize an EntityMux.
The definitions are expected to be an array of struct Types
which define the Entities to be used in this application.
There is a set of tags available to decorate fields in these
struct Types to configure the EntityMux.

It firstly creates the appropriate collection in the database
for every definition using information from the EntityID field.
The collection is packaged, along with the Type definition for
every struct and stored in the EntityMux.
*/
func Create(db *mongo.Database, definitions []interface{}) (*EntityMux, error) {
	newMux := &EntityMux{entities: map[string]*entity.Entity{}}

	for def := range definitions {
		defType := reflect.TypeOf(def)

		var collectionInitialized bool
		/*
			This block performs a linear search through the fields of this
			def to find the first one with a non-empty IDTag.
			The field's IDTag's value is used as the unique identifier
			for this def's Entity.

			For efficiency, clients should always put the IDTag as
			the first field in a struct.
		*/
		for i := 0; i < defType.NumField(); i++ {
			field := defType.Field(i)

			if tag := field.Tag.Get(entity.IDTag); tag != "" {
				collectionName := tag

				var defCollection *mongo.Collection

				if collectionOptions := collectionValidator(defType); collectionOptions != nil {
					defCollection = db.Collection(collectionName, collectionOptions)
				} else {
					defCollection = db.Collection(collectionName)
				}

				defEntity := &entity.Entity{
					SchemaDefinition: entity.TypeOf(def),
					PStorage:         defCollection,
				}

				if newMux.entities[collectionName] == nil {
					newMux.entities[collectionName] = defEntity
				} else {
					return nil, fmt.Errorf("duplicate '%s' tag on '%s'",
						entity.IDTag, defType.Name())
				}

				defEntity.Optimize()

				collectionInitialized = true
				break
			}
		}

		if !collectionInitialized {
			return nil, fmt.Errorf("no '%s' tag on '%s'",
				entity.IDTag, defType.Name())
		}
	}

	return newMux, nil
}

/*
collectionValidator uses the given Type to access field
tags for the struct and create a collection validator to
be used for this struct.
*/
func collectionValidator(_ reflect.Type) *options.CollectionOptions {
	// TODO: implement collection validator
	return nil
}
