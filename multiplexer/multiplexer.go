/*
Package multiplexer defines an EntityMux type which is basically
a multiplexer for Entity types.
It uses struct field tags in Entity
*/
package multiplexer

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/navaz-alani/entity"
	"github.com/navaz-alani/entity/multiplexer/eMuxContext"
	"github.com/navaz-alani/entity/pkgErrors"
)

/*
condensedField is a shorthand representation of the information
which is commonly used/is important within the context of this
package.
*/
type condensedField struct {
	Name string
	Type reflect.Type
	/*
		RequestID is a string which specifies the fieldName to
		expect when parsing JSON for this field.
	*/
	RequestID   string
	StructIndex []int
}

/*
metaEntity is a parsed version of an Entity's
struct tags.
It stores information that is frequently needed.
*/
type metaEntity struct {
	Entity *entity.Entity
	/*
		EntityID is the value of the first non-empty
		entity.IDTag value.
	*/
	EntityID string
	/*
		FieldClassifications maps a field classification to
		an array of pointers of condensedFields.
	*/
	FieldClassifications *map[rune][]*condensedField
}

/*
The following constants define the rune keys to the field
classifications in a metaEntity.
*/
const (
	/*
		classCollectionName maps to an array containing a
		single (or none at all) pointer to a condensedField
		whose RequestID is the Entity's collectionName in the
		database.
	*/
	classCollectionName rune = '*'
	classCreationField  rune = 'c'
)

/*
EntityMux is a multiplexer for Entities.
*/
type EntityMux struct {
	/*
		entities is a collection of Entities which are
		used in an application.
		In the entities map, the key is the EntityID.
	*/
	entities map[string]*metaEntity
	/*
		Router is the httprouter.Router on which the
		EntityMux configuration has been set-up; other
		than that, it's just a new httprouter.Router.

		Once an EntityMux has been created, the client
		can use this pointer to further configure the router,
		and then serve at a later point.
	*/
	Router *httprouter.Router
}

/*
Collection returns a pointer to the mongo collection
that the entity identified by the given entityID is
using.
*/
func (em *EntityMux) Collection(entityID string) *mongo.Collection {
	return em.entities[entityID].Entity.PStorage
}

/*
CreationMiddleware returns httprouter middleware which can be used to
derive a template of an Entity from an API request.
It uses the HandleTag as well as the JSON/BSON/field name (in that
priority) to to determine what to search for in incoming JSON
payloads.

TODO: document and implement linking of entities
NOTE: This functionality does not yet support embedding of Entity
types. This can be achieved through linking instead.
*/
func (em *EntityMux) CreationMiddleware(entityID string) (func(next httprouter.Handle) httprouter.Handle, error) {
	var thisEntity *metaEntity
	if meta := em.entities[entityID]; meta.EntityID == "" {
		return nil, pkgErrors.IncompleteEntityMetadata
	} else {
		thisEntity = meta
	}

	handle := func(next httprouter.Handle) httprouter.Handle {
		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			// Decode the incoming JSON payload
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "payload decode fail", http.StatusBadRequest)
				return
			}

			/*
				This block processes each of the creation fields for this
				Entity and copies their values in the request payload to a
				reflect.Value which will be serialized as an interface, ready
				for the client to retrieve through a type assertion.

				If the type assertion fails, the request data is probably
				malformed and the client can handle this.
			*/
			entityType := thisEntity.Entity.SchemaDefinition
			encodedEntityValue := reflect.New(entityType)

			creationFields := (*thisEntity.FieldClassifications)[classCreationField]
			if creationFields == nil {
				creationFields = []*condensedField{}
			}

			// Compare the fields in the struct
			for i := 0; i < len(creationFields); i++ {
				field := creationFields[i]

				// check that the payload contains this field
				if fieldVal := req[field.RequestID]; fieldVal != "" {
					f := encodedEntityValue.FieldByIndex(field.StructIndex).Elem()
					if !f.CanSet() {
						continue
					}

					/*
						Interface serialised to bytes.
						Client decodes bytes to interface, which is type asserted
						against the client's implementation.
					*/
					var serializedValue bytes.Buffer
					enc := gob.NewEncoder(&serializedValue)
					err := enc.Encode(fieldVal)
					if err != nil {
						continue
					}

					f.SetBytes(serializedValue.Bytes())
				}
			}

			// place encodedEntityValue in the request context
			muxCtx := eMuxContext.EMuxContext{}
			muxCtx.PackagePayload(eMuxContext.EMuxKey, encodedEntityValue)

			next(w, muxCtx.ContextualizedRequest(r, context.Background(), eMuxContext.EMuxKey), ps)
		}
	}

	return handle, nil
}

/*
requestID returns the fieldName to expect when parsing for this
field in incoming request JSON payloads.
*/
func requestID(field *reflect.StructField) string {
	if tag := field.Tag.Get(entity.JSONTag); tag != "" {
		return tag
	} else if tag := field.Tag.Get(entity.BSONTag); tag != "" {
		return tag
	} else {
		return field.Name
	}
}

/*
classifyFields is a function which iterates over the fields of
the given Type and classifies them by their HandleTags.
*/
func classifyFields(defType reflect.Type) map[rune][]*condensedField {
	classifications := map[rune][]*condensedField{}

	collectionName := classifications[classCollectionName]
	creationFields := classifications[classCreationField]

	for i := 0; i < defType.NumField(); i++ {
		field := defType.Field(i)
		fieldCondensed := &condensedField{
			Name:        field.Name,
			Type:        field.Type,
			RequestID:   requestID(&field),
			StructIndex: field.Index,
		}

		if tag := field.Tag.Get(entity.IDTag); (collectionName == nil || len(collectionName) == 0) &&
			tag != "" {
			collectionName = []*condensedField{fieldCondensed}
		}

		if tag := field.Tag.Get(entity.HandleTag); strings.ContainsAny(tag, "c") {
			if creationFields == nil || len(creationFields) == 0 {
				creationFields = []*condensedField{fieldCondensed}
			} else {
				creationFields = append(creationFields, fieldCondensed)
			}
		}
	}

	return classifications
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
	newMux := &EntityMux{entities: map[string]*metaEntity{}}

	for i := 0; i < len(definitions); i++ {
		defType := reflect.TypeOf(definitions[i])
		fieldClassifications := classifyFields(defType)

		var collectionName string
		collectionNameClassification := fieldClassifications[classCollectionName]

		if collectionNameClassification == nil || len(collectionNameClassification) == 0 ||
			collectionNameClassification[0].RequestID == "" {
			return nil, fmt.Errorf("no '%s' tag on '%s'",
				entity.IDTag, defType.Name())
		} else {
			collectionName = collectionNameClassification[0].RequestID
		}

		var defCollection *mongo.Collection
		if collectionOptions := collectionValidator(defType); collectionOptions != nil {
			defCollection = db.Collection(collectionName, collectionOptions)
		} else {
			defCollection = db.Collection(collectionName)
		}

		defEntity := &entity.Entity{
			SchemaDefinition: defType,
			PStorage:         defCollection,
		}

		if newMux.entities[collectionName] == nil {
			newMux.entities[collectionName] = &metaEntity{
				Entity:               defEntity,
				EntityID:             collectionName,
				FieldClassifications: &fieldClassifications,
			}
		} else {
			return nil, fmt.Errorf("duplicate '%s' tag on '%s'",
				entity.IDTag, defType.Name())
		}

		defEntity.Optimize()
		break
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
