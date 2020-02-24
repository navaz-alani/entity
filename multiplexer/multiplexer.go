/*
Package multiplexer defines an EntityMux type which is basically
a multiplexer for Entity types.
It uses struct eField tags in Entity definitions in order to
create database collections, middleware for request pre-processing
and more.

Tags

Here are the eField tags that the EntityMux uses:

entity.IDTag - This tag is used to give a name to an Entity.
This name specifies the mongo.Collection that will be created
in the database for an Entity. It is also used by EntityMux to
internally work with Entity types. This value must be unique
amongst the Entity types that the EntityMux manages.

entity.HandleTag - This tag is used to provide configurations
for middleware generation. The value for this tag is a string
containing configuration tokens. These tokens are single characters
(runes) which can be used to classify a eField. For example, the
CreationFieldsToken token can be used used to specify which
fields should be parsed from an http.Response body for the
middleware generation.

entity.AxisTag - This tag is used to specify which fields can be
considered to be unique (to an Entity) within a collection.
The tag value which indicates that a eField is an axis eField is
the string "true"-- all other values are rejected.

entity.IndexTag - This tag is used to specify the fields for which
an index needs to be built in the database collection. This is used
hand in hand with the entity.Axis tag; in order for a eField's index
to be constructed, both these tags have to be set to "true".
*/
package multiplexer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/navaz-alani/entity"
	"github.com/navaz-alani/entity/entityErrors"
	"github.com/navaz-alani/entity/multiplexer/muxContext"
	"github.com/navaz-alani/entity/multiplexer/muxHandle"
)

/*
EntityMux is a multiplexer for Entities.
It is designed to manage initialization for multiple
Entity types from their struct definitions.

See the Create function for more information about
EntityMux initialization steps.
*/
type EntityMux struct {
	/*
		Entities is a collection of Entities which are
		used in an application.
		In the Entities map, the key is the EntityID.
	*/
	Entities map[string]*metaEntity
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
using for persistent storage.

To modify the options for the collection, the client can
use the db pointer used during initialization
*/
func (em *EntityMux) Collection(entityID string) *mongo.Collection {
	return em.Entities[entityID].Entity.PStorage
}

/*
E returns the Entity corresponding to the entityID given.
*/
func (em *EntityMux) E(entityID string) *entity.Entity {
	if meta := em.Entities[entityID]; meta != nil {
		return meta.Entity
	}
	return nil
}

/*
Create uses the given definitions to create an EntityMux which manages the
corresponding Entities. The definitions are expected to be an array of
empty/zero struct Types. For example, consider the User entity defined in
the "Getting Started" section of the documentation of the entity package.
In order to create an EntityMux which manages the User Entity, the following
line suffices:

	eMux, err := multiplexer.Create(dbPtr, User{})

When internally registering Entities, a unique identifier is needed to
refer to Entities. This identifier is called the EntityID and is defined using
the IDTag. If the IDTag is not defined for any entity, the multiplexer may not
be correctly initialized and an entityErrors.IncompleteEntityMetadata is
returned.

Remember, when instantiating an Entity, it is important to have a defined
location for persistent storage. In this case, it is a *mongo.Collection.
For each definition, a collection in the database is initialized. The name of
this collection is exactly the same as the definition's EntityID.

When initializing the collection, a schema validator is first created. If this
is successful, the validator is injected as an option when creating the collection.
Otherwise, the collection is created without a
The validator also uses tags to generate schemas for validation. For more
information, see the CollectionValidator function.

After each collection has been created and linked to the respective Entity,
the Entity's Optimize() method is called to index the axis fields which have
been marked for indexing.
*/
func Create(db muxHandle.DBHandle, definitions ...interface{}) (*EntityMux, error) {
	if db == nil {
		return nil, entityErrors.DBUninitialized
	}

	entityMap := make(map[string]*metaEntity)
	newMux := &EntityMux{Entities: entityMap}

	for i := 0; i < len(definitions); i++ {
		defType := reflect.TypeOf(definitions[i])
		fieldClassifications := classifyFields(defType)

		var collectionName string
		collectionNameClassification := fieldClassifications[CollectionIDToken]

		if len(collectionNameClassification) == 0 || collectionNameClassification[0].Value == "" {
			return nil, entityErrors.NoTag(entity.IDTag, defType.Name())
		} else {
			collectionName, _ = collectionNameClassification[0].Value.(string)
		}

		var defCollection *mongo.Collection
		if collectionOptions := CollectionValidator(defType); collectionOptions != nil {
			defCollection = db.Collection(collectionName, collectionOptions)
		} else {
			defCollection = db.Collection(collectionName)
		}

		defEntity := &entity.Entity{
			SchemaDefinition: defType,
			PStorage:         defCollection,
		}

		if newMux.Entities[collectionName] == nil {
			newMux.Entities[collectionName] = &metaEntity{
				Entity:               defEntity,
				EntityID:             collectionName,
				FieldClassifications: fieldClassifications,
			}
		} else {
			return nil, entityErrors.DuplicateTag(entity.IDTag, defType.Name())
		}

		_ = defEntity.Optimize()
	}

	return newMux, nil
}

/*
CreationMiddleware returns httprouter middleware which can be used to
derive a template of an Entity from an API request.

The creation fields for the Entity corresponding to the given entityID
are used to pre-populate the response context with an "auto-filled"
Entity.
For each creation eField, the first non-empty value of JSON/BSON/eField name
is used to check the incoming request payload for a corresponding value.
This means that if the JSONTag is defined for the eField, it will be assumed
to be the corresponding eField in the JSON payload. Otherwise, the BSONTag
is checked next. If the BSONTag is also empty, the eField's name is used.

The returned function is middleware which can be used on an httprouter.Router
so that when a request is received by the client's httprouter.DBHandle, an
auto-completed version of the entity is present in the request context.

NOTE: This functionality does not yet support embedding of Entity
types. This can be achieved through linking instead. This is a
feature which has been planned for implementation.
*/
func (em *EntityMux) CreationMiddleware(entityID string) (func(next httprouter.Handle) httprouter.Handle, error) {
	var meta *metaEntity
	if m := em.Entities[entityID]; m.EntityID == "" {
		return nil, entityErrors.IncompleteEntityMetadata
	} else {
		meta = m
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
			entityType := meta.Entity.SchemaDefinition
			preProcessedEntity := reflect.New(entityType)

			creationFields := meta.FieldClassifications[CreationFieldsToken]

			// Compare the fields in the struct
			for i := 0; i < len(creationFields); i++ {
				field := creationFields[i]

				if fieldVal := req[field.RequestID]; fieldVal != "" {
					f := preProcessedEntity.Elem().FieldByName(field.Name)
					if !f.CanSet() {
						continue
					}
					_ = em.writeToField(f, fieldVal)
				}
			}

			muxCtx := muxContext.Create()
			muxCtx.Set(meta.EntityID, preProcessedEntity.Interface())

			reqWithCtx := muxCtx.ContextualizeRequest(r, context.Background(), muxContext.EMuxKey)
			next(w, reqWithCtx, ps)
		}
	}

	return handle, nil
}

/*
writeToField takes a eField value and attempts to set
its value to the given data.
This function will NEVER write to a eField which stores
a pointer kind.
*/
func (em *EntityMux) writeToField(field reflect.Value, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	/*
		Do not need to support pointers because an Entity has database handles.
		Pointers stored in databases would make no sense and therefore there is
		no pointer case in this switch.
	*/
	switch field.Kind() {
	case reflect.String:
		field.SetString(data.(string))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(data.(int64))
	case reflect.Float32, reflect.Float64:
		field.SetFloat(data.(float64))
	case reflect.Bool:
		field.SetBool(data.(bool))
	default:
		field.Set(reflect.ValueOf(data))
	}

	/*
		TODO: embedding entities (read below)
		Loop through the existing entities and check for an
		embedded eField. Use that eField's type in order to
		determine the embedded Entity.
		The use eField.Set(reflect.ValueOf(data))
	*/

	return nil
}

/*
CollectionValidator uses the given Type to access eField
tags for the struct and create a collection validator to
be used for this struct.

TODO: This function still has to be implemented.
*/
func CollectionValidator(_ reflect.Type) *options.CollectionOptions {
	return nil
}
