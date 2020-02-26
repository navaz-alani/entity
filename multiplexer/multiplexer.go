package multiplexer

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"reflect"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/navaz-alani/entity"
	"github.com/navaz-alani/entity/eField"
	"github.com/navaz-alani/entity/entityErrors"
	"github.com/navaz-alani/entity/multiplexer/muxContext"
	"github.com/navaz-alani/entity/multiplexer/muxHandle"
)

/*
EntityMux is a multiplexer for Entities.
It is meant to manage multiple Entities within an application.
This involves creating and linking a database collection for
Entities, generating pre-processing middleware for CRUD requests,
and verification.

See the Create function for more information about the
EntityMux initialization.
*/
type (
	EntityMux struct {
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
func (em *EntityMux) Collection(entityID string) *mongo.Collection {
	return em.Entities[entityID].Entity.PStorage
}

/*
E returns the Entity corresponding to the entityID given.

This Entity can be used normally to carry out CRUD operations
for instances of the Entity.
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
For each definition, a collection in the database is initialized iff the IDTag
does NOT start with a "!". The name of the collection created is exactly the
same as the definition's EntityID (last IDTag value). Note, also, that the "!"
used when avoiding collection creation does NOT could as part of the EntityID.

Entities for which a database collection has been created are then indexed
against their axis fields which have been marked for indexing. A field can be
specified as an axis field by using the entity.AxisTag while index creation is
specified using the entity.IndexTag. Only fields with the AxisTag set to "true"
and a non-empty IndexTag are indexed.

When initializing the collection, a schema validator is first created. If this
is successful, the validator is injected as an option when creating the collection.
Otherwise, the collection is created without a
The validator also uses tags to generate schemas for validation. For more
information, see the CollectionValidator function.
*/
func Create(db muxHandle.DBHandler, definitions ...interface{}) (*EntityMux, error) {
	if db == nil {
		return nil, entityErrors.DBUninitialized
	}

	entityMap := make(map[string]*metaEntity)
	typeMap := make(map[reflect.Type]string)
	newMux := &EntityMux{Entities: entityMap, TypeMap: typeMap}

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
			if collectionOptions := CollectionValidator(defType); collectionOptions != nil {
				defCollection = db.Collection(EntityID, collectionOptions)
			} else {
				defCollection = db.Collection(EntityID)
			}
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
func (em *EntityMux) link() {
	for _, meta := range em.Entities {
		// todo: append other field classes to `fields` for linking too
		fields := meta.FieldClassifications[CreationFieldsToken]

		for i := 0; i < len(fields); i++ {
			field := fields[i]

			var embedID string
			switch field.Type.Kind() {
			case reflect.Slice:
				embedID = em.TypeMap[field.Type.Elem()]
			case reflect.Struct:
				embedID = em.TypeMap[field.Type]
			}

			if embedID == "" {
				continue
			}

			// create reference to embedded Entity metadata.
			field.EmbeddedEntity = em.Entities[embedID]
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
func (em *EntityMux) CreationMiddleware(entityID string) (func(next http.Handler) http.Handler, error) {
	var meta *metaEntity
	if m := em.Entities[entityID]; m.EntityID == "" {
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

			preProcessedEntity, err := em.processCreationPayload(em.Entities[entityID], req)
			if err != nil {
				// JSON pre-processing failed
				//		TODO: add error in context for inspection purposes
				next.ServeHTTP(w, r)
				return
			}

			muxCtx := muxContext.Create()
			_ = muxCtx.Set(meta.EntityID, &preProcessedEntity)

			reqWithCtx := muxCtx.EmbedCtx(r, context.Background())
			next.ServeHTTP(w, reqWithCtx)
		})
	}

	return handle, nil
}

/*
CollectionValidator uses the given Type to access eField
tags for the struct and create a collection validator to
be used for this struct.

It basically creates a JSON schema for the given type.
TODO: This function still has to be implemented.
*/
func CollectionValidator(_ reflect.Type) *options.CollectionOptions {
	/*
		Implementation Details:

		Use tags to populate JSON schema for each field of the current type.
		When an embedded field is reached, recursively compute the JSON schema.
		A base case is guaranteed if types are not recursive.

		JSON schema fields to populate: BSON type, required
	*/
	return nil
}

/*
processCreationPayload parses the given JSON payload with respect to the
entity corresponding to the given entityID.
*/
func (em *EntityMux) processCreationPayload(meta *metaEntity, payload map[string]interface{}) (reflect.Value, error) {
	var preProcessedEntity reflect.Value
	var creationFields []*condensedField

	if meta == nil {
		return reflect.ValueOf(nil), entityErrors.InvalidEntityID
	} else {
		preProcessedEntity = reflect.New(meta.Entity.SchemaDefinition)
		creationFields = meta.FieldClassifications[CreationFieldsToken]
	}

	/*
		This block processes each of the creation fields for this
		Entity and copies their values in the request payload to a
		reflect.Value which will be serialized as an interface, ready
		for the client to retrieve through a type assertion.

		If the type assertion fails, the request data is probably
		malformed and the client can handle this.

		TODO: in json int is parsed as float64; needs to be handled
	*/
	for i := 0; i < len(creationFields); i++ {
		field := creationFields[i]

		if fieldVal := payload[field.RequestID]; fieldVal != "" {
			// check write status
			f := preProcessedEntity.Elem().FieldByName(field.Name)
			if !f.CanSet() {
				continue
			}

			data := fieldVal

			if field.EmbeddedEntity != nil {
				// todo: extract embedded payload
				embeddedPayload, ok := fieldVal.(map[string]interface{})
				if !ok {
					log.Println("embedded payload invalid")
					continue
				}

				embedValue, err := em.processCreationPayload(field.EmbeddedEntity, embeddedPayload)
				if err != nil {
					log.Println(err)
					continue
				}
				data = embedValue.Elem().Interface()
			}

			err := writeToField(&f, data)
			switch err {
			case entityErrors.InvalidDataType:
				log.Println(err)
				continue
			}
		}
	}

	return preProcessedEntity, nil
}
