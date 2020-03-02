package multiplexer

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/navaz-alani/entity"
	"github.com/navaz-alani/entity/eField"
	"github.com/navaz-alani/entity/entityErrors"
	"github.com/navaz-alani/entity/multiplexer/muxContext"
	"github.com/navaz-alani/entity/multiplexer/muxHandle"
)

type (
	/*
		EMux is a multiplexer for Entities.
		It is meant to manage multiple Entities within an application.
		This involves creating and linking a database collection for
		Entities, generating pre-processing middleware for CRUD requests,
		and verification.

		See the Create function for more information about the
		EMux initialization.
	*/
	EMux struct {
		/*
			Entities is a collection of Entities which are
			used in an application.
			In the Entities map, the key is the EntityID.
		*/
		Entities EntityMap
		/*
			TypeMap provides a way of performing a reverse
			lookup for EntityID by a reflect.Type
		*/
		TypeMap TypeMap
	}

	/*
		EntityMap is a type used to store Entities for look-up by
		their EntityID.
	*/
	EntityMap map[string]*metaEntity
	/*
		TypeMap is a type used to perform a reverse lookup for an
		Entity by type of an instance.
	*/
	TypeMap map[reflect.Type]string
)

/*
Collection returns a pointer to the underlying mongo.Collection
that the entity corresponding to the given entityID is using for
persistent storage.

Alternatively, a handle to this collection can be obtained by
using the go.mongodb.org/mongo-driver directly with the collection
name as the EntityID given.

To modify the options for the collection, the client can
use the db pointer used during initialization
*/
func (em *EMux) Collection(entityID string) *mongo.Collection {
	return em.Entities[entityID].Entity.PStorage
}

/*
E returns the Entity corresponding to the entityID given.

This Entity can be used normally to carry out CRUD operations
for instances of the Entity.
*/
func (em *EMux) E(entityID string) *entity.Entity {
	if meta := em.Entities[entityID]; meta != nil {
		return meta.Entity
	}
	return nil
}

/*
Create uses the given definitions to create an EMux which manages the
corresponding Entities. The definitions are expected to be an array of
empty/zero struct Types. For example, consider the User entity defined in
the "Getting Started" section of the documentation of the entity package.
In order to create an EMux which manages the User Entity, the following
line suffices:

	eMux, err := multiplexer.Create(dbPtr, User{})

When internally registering Entities, a unique identifier is needed to
refer to Entities. This identifier is called the EntityID and is defined using
the IDTag. If the IDTag is not defined for any entity, the multiplexer may not
be correctly initialized and an entityErrors.IncompleteEntityMetadata is
returned.

Remember, when instantiating an Entity, it is important to have a defined
location for persistent storage. In this case, it is a *mongo.Collection.
For each definition, a collection in the database is initialized iff the IDTag
does NOT start with a "!". The name of the collection created is exactly the
same as the definition's EntityID (last IDTag value). Note, also, that the "!"
used when avoiding collection creation does NOT could as part of the EntityID.

Entities for which a database collection has been created are then indexed
against their axis fields which have been marked for indexing. A field can be
specified as an axis field by using the entity.AxisTag while index creation is
specified using the entity.IndexTag. Only fields with the AxisTag set to "true"
and a non-empty IndexTag are indexed.
*/
func Create(db muxHandle.DBHandler, definitions ...interface{}) (*EMux, error) {
	if db == nil {
		return nil, entityErrors.DBUninitialized
	}

	entityMap := make(map[string]*metaEntity)
	typeMap := make(map[reflect.Type]string)
	newMux := &EMux{Entities: entityMap, TypeMap: typeMap}

	// populate entity metadata
	for i := 0; i < len(definitions); i++ {
		defType := reflect.TypeOf(definitions[i])
		fieldClassifications := classifyFields(defType)

		createCollection := true
		var EntityID string

		// Extract collection name
		collectionNameClassification := fieldClassifications[EntityIDToken]
		if len(collectionNameClassification) == 0 || collectionNameClassification[0].Value == "" {
			return nil, entityErrors.NoTag(eField.IDTag, defType.Name())
		} else if collectionNameClassification[0].Value[0] != '!' {
			EntityID = collectionNameClassification[0].Value
		} else {
			EntityID = collectionNameClassification[0].Value[1:]
			createCollection = false
		}

		// create collection
		var defCollection *mongo.Collection
		if createCollection {
			defCollection = db.Collection(EntityID)
		}

		// create & register entity
		defEntity := &entity.Entity{
			SchemaDefinition: defType,
			PStorage:         defCollection,
		}

		if newMux.Entities[EntityID] == nil {
			meta := &metaEntity{
				Entity:               defEntity,
				EntityID:             EntityID,
				FieldClassifications: fieldClassifications,
			}

			newMux.Entities[EntityID] = meta
			newMux.TypeMap[defType] = EntityID
		} else {
			return nil, entityErrors.DuplicateTag(eField.IDTag, defType.Name())
		}

		// run indexing
		if EntityID != "" {
			_ = defEntity.Optimize()
		}
	}

	newMux.link()
	return newMux, nil
}

/*
link creates internal representations of embedded struct field types
for parsing in middleware.
*/
func (em *EMux) link() {
	for _, meta := range em.Entities {
		// todo: append other field classes to `fields` for linking too
		fields := meta.FieldClassifications[CreationFieldsToken]

		for i := 0; i < len(fields); i++ {
			field := fields[i]

			var embedID string
			if field.EmbeddedEntity.CFlag || field.EmbeddedEntity.SFlag {
				embedID = em.TypeMap[field.EmbeddedEntity.EmbeddedType]
			} else {
				embedID = em.TypeMap[field.Type]
			}

			if embedID == "" {
				continue
			}

			// create reference to embedded Entity metadata
			field.EmbeddedEntity.Meta = em.Entities[embedID]
		}
	}
}

/*
CreationMiddleware returns middleware which can be used to
derive a template of an Entity/CRUD operation from an API request.

The creation fields for the Entity corresponding to the given entityID
are used to pre-populate the response context with an pre-populated"
Entity.

For each creation eField, the first non-empty value of JSON/BSON/eField name
is used to check the incoming request payload for a corresponding value.
This means that if the JSONTag is defined for the eField, it will be assumed
to be the corresponding eField in the JSON payload. Otherwise, the BSONTag
is checked next. If the BSONTag is also empty, the eField's name is used.

The returned function is middleware which can be used on an httprouter.Router
so that when a request is received by the client's httprouter.DBHandler, an
auto-completed version of the entity is present in the request context.

NOTE: This functionality does not yet support embedding of Entity
types. This can be achieved through linking instead. This is a
feature which has been planned for implementation.
*/
func (em *EMux) CreationMiddleware(entityID string) (func(next http.Handler) http.Handler, error) {
	var meta *metaEntity
	if m := em.Entities[entityID]; m == nil || m.EntityID == "" {
		return nil, entityErrors.IncompleteEntityMetadata
	} else {
		meta = m
	}

	if len(meta.FieldClassifications[CreationFieldsToken]) == 0 {
		return nil, entityErrors.NoClassificationFields
	}

	handle := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Decode the incoming JSON payload
			var req map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "payload decode fail", http.StatusBadRequest)
				return
			}

			preProcessedEntity, err := em.createEntity(em.Entities[entityID], req)
			if err != nil {
				// JSON pre-processing failed
				//		TODO: add error in context for inspection purposes
				next.ServeHTTP(w, r)
				return
			}

			muxCtx := muxContext.Create()
			_ = muxCtx.Set(meta.EntityID, preProcessedEntity.Interface())

			reqWithCtx := muxCtx.EmbedCtx(r, context.Background())
			next.ServeHTTP(w, reqWithCtx)
		})
	}

	return handle, nil
}

func (em *EMux) createEntity(meta *metaEntity, payload map[string]interface{}) (reflect.Value, error) {
	var preProcessedEntity reflect.Value
	var creationFields []*condensedField

	if meta == nil {
		return reflect.ValueOf(nil), entityErrors.InvalidEntityID
	} else {
		preProcessedEntity = reflect.New(meta.Entity.SchemaDefinition).Elem()
		creationFields = meta.FieldClassifications[CreationFieldsToken]
	}

	for _, cf := range creationFields {
		// check if there is data to be written to this field
		if fieldData := payload[cf.RequestID]; fieldData != nil {
			fieldToWrite := preProcessedEntity.FieldByName(cf.Name)

			if cf.EmbeddedEntity.CFlag {
				if cf.EmbeddedEntity.Meta == nil {
					return preProcessedEntity, entityErrors.InvalidEntityLink
				}

				// convert field's payload to slice of interfaces
				writeData, ok := fieldData.([]interface{})
				if !ok {
					return preProcessedEntity, entityErrors.EmbeddedWriteDataInvalid
				}

				// write each item individually
				for i := 0; i < len(writeData); i++ {
					writeItem := writeData[i]

					// convert payload for recursive call
					writeMap, ok := writeItem.(map[string]interface{})
					if !ok {
						return preProcessedEntity, entityErrors.EmbeddedWriteDataInvalid
					}

					// recursively create entity for field
					writeValue, err := em.createEntity(cf.EmbeddedEntity.Meta, writeMap)
					if err != nil {
						return preProcessedEntity, err
					}

					// append new value
					fieldToWrite.Set(reflect.Append(fieldToWrite, writeValue))
				}
				continue
			} else if cf.EmbeddedEntity.SFlag {
				// convert payload for recursive call
				writeData, ok := fieldData.(map[string]interface{})
				if !ok {
					return preProcessedEntity, entityErrors.EmbeddedWriteDataInvalid
				}

				// recursively create entity for field
				embedValue, err := em.createEntity(cf.EmbeddedEntity.Meta, writeData)
				if err != nil {
					return preProcessedEntity, entityErrors.EmbeddedWriteDataInvalid
				}

				// set data to be written
				fieldData = embedValue.Interface()
			}

			// set data
			if err := eField.WriteToField(&fieldToWrite, fieldData); err != nil {
				return preProcessedEntity, err
			}
		}
	}

	return preProcessedEntity, nil
}
