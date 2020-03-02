package multiplexer

import (
	"reflect"
	"strings"

	"github.com/navaz-alani/entity"
	"github.com/navaz-alani/entity/eField"
)

type (
	/*
	   metaEntity is a parsed version of an Entity's
	   struct tags.
	   It stores information that is frequently needed.
	*/
	metaEntity struct {
		Entity *entity.Entity
		/*
			EntityID is the value of the first non-empty
			entity.IDTag value.
		*/
		EntityID string
		/*
			FieldClassifications maps a field classification to
			a slice of pointers to condensedFields.
		*/
		FieldClassifications map[rune][]*condensedField
	}

	/*
	   condensedField is a shorthand representation of the information
	   which is commonly used/is important within the context of this
	   package.
	*/
	condensedField struct {
		// Name is the field's field name.
		Name string
		// Type is the reflection type of the field.
		Type reflect.Type
		/*
			RequestID is a string which specifies the field to
			expect when parsing JSON for this field.
			It can be equal to the Name field.
		*/
		RequestID string
		/*
			Value is used to store the collection name that this
			field specifies.
		*/
		Value string
		/*
			EmbeddedEntity is used to store an internal reference to
			the Entity whose type this field specifies.
		*/
		EmbeddedEntity Embedding
	}

	/*
		Embedding is a type used to store information about a field's
		data type. It contains flags used to indicate whether the field
		type is another Entity managed by the multiplexer or whether
		it is a collection-kind field.
	*/
	Embedding struct {
		/*
			SFlag is a boolean representing whether a field
			stores an internally managed Entity (kind struct).
		*/
		SFlag bool
		/*
			CFlag is a boolean representing whether a field
			stores collection-type data (slice, array, ...)
		*/
		CFlag bool
		/*
			EmbeddedType specifies the field's embedded type.
		*/
		EmbeddedType reflect.Type
		/*
			Meta is a pointer representing an internal link
			to an internally managed Entity.
		*/
		Meta *metaEntity
	}
)

/*
   The following constants define the tokens for field
   classifications in a metaEntity. These tokens are to be
   used within the context of the entity.HandleTag.

	Note: "request payload" refers to a JSON payload.
*/
const (
	/*
		EntityIDToken maps to a slice containing a
		single pointer (or none at all) to a condensedField
		whose RequestID is the Entity's EntityID. This is
		used within the creation stage.
	*/
	EntityIDToken rune = '*'
	/*
		AxisFieldToken maps to a slice containing fields which
		are tagged as axis fields.
	*/
	AxisFieldToken rune = 'a'
	/*
		CreationFieldsToken maps to a slice containing fields
		which were specified to be provided in an http.Request
		for creating an instance of an Entity.
	*/
	CreationFieldsToken rune = 'c'
	/*
		DeletionFieldsToken maps to a slice containing fields
		which were specified to be pre-processed from a delete
		request's payload.
	*/
	DeletionFieldsToken rune = 'd'
	/*
		EditFieldsToken maps to a slice containing fields which
		were specified to be editable from an http.Request.
	*/
	EditFieldsToken rune = 'e'
)

/*
HandleTokens defines the set of tokens which can be used in
the entity.HandleTag of a struct field for classification.
*/
var HandleTokens = []rune{
	AxisFieldToken, // uniqueness
	// operation tokens
	CreationFieldsToken, // creation
	DeletionFieldsToken, // deletion
	EditFieldsToken,     // editing
}

/*
classifyFields is a function which iterates over the fields of
the given Type and classifies them by their HandleTag tokens.
*/
func classifyFields(defType reflect.Type) map[rune][]*condensedField {
	classifications := map[rune][]*condensedField{}

	for i := 0; i < defType.NumField(); i++ {
		classifyHandleTags(defType.Field(i), classifications)
	}

	return classifications
}

/*
classifyHandleTags classifies the given field by its handle tags.
For every tag that the field matches, a pointer to a condensedField
representation of the given field is added to the corresponding
tag's field array in the given class map.

If an entity.IDTag is encountered, the collectionID is reset. This
means that the last entity.IDTag will specify the value of the
entity's mongoDB collection.
*/
func classifyHandleTags(field reflect.StructField, classes map[rune][]*condensedField) {
	cFlag, sFlag, embeddedType := eField.CheckEmbedding(field)

	newField := &condensedField{
		Name:      field.Name,
		Type:      field.Type,
		RequestID: eField.NameByPriority(field, eField.PriorityJsonBson),
		EmbeddedEntity: Embedding{
			CFlag:        cFlag,
			SFlag:        sFlag,
			EmbeddedType: embeddedType,
		},
	}

	for _, tok := range HandleTokens {
		if classes[tok] == nil {
			classes[tok] = make([]*condensedField, 0)
		}

		if tag := field.Tag.Get(eField.IDTag); tag != "" && tag != "-" {
			/*
				No need to check if tag starts with "!" because that will
				be done in the creation stage.
			*/
			newField.Value = tag
			classes[EntityIDToken] = make([]*condensedField, 0)
			classes[EntityIDToken] = append(classes[EntityIDToken], newField)
		}

		if tag := field.Tag.Get(eField.HandleTag); strings.ContainsAny(tag, string(tok)) {
			classes[tok] = append(classes[tok], newField)
		}
	}
}
