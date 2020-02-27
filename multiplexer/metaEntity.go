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
			FieldClassifications maps a eField classification to
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
		// Name is the eField's eField name.
		Name string
		// Type is the reflection type of the eField.
		Type reflect.Type
		/*
			RequestID is a string which specifies the eField to
			expect when parsing JSON for this eField.
			It can be equal to the Name eField.
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
		EmbeddedEntity *metaEntity
	}
)

/*
   The following constants define the tokens for eField
   classifications in a metaEntity. These tokens are to be
   used within the context of the entity.HandleTag.
*/
const (
	/*
		EntityIDToken maps to an array containing a
		single pointer (or none at all) to a condensedField
		whose RequestID is the Entity's EntityID. This is
		used within the creation stage.
	*/
	EntityIDToken rune = '*'
	/*
		AxisFieldToken maps to an array containing fields which
		are tagged as axis fields.
	*/
	AxisFieldToken rune = 'a'
	/*
		CreationFieldsToken maps to an array containing fields
		which were specified to be provided in an http.Request
		for creating an instance of an Entity.
	*/
	CreationFieldsToken rune = 'c'
)

/*
HandleTokens defines the set of tokens which can be used in
the entity.HandleTag of a struct eField for classification.
*/
var HandleTokens = []rune{
	CreationFieldsToken,
	AxisFieldToken,
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
classifyHandleTags classifies the given eField by its handle tags.
For every tag that the eField matches, a pointer to a condensedField
representation of the given eField is added to the corresponding
tag's eField array in the given class map.

If an entity.IDTag is encountered, the collectionID is reset. This
means that the last entity.IDTag will specify the value of the
entity's mongoDB collection.
*/
func classifyHandleTags(field reflect.StructField, classes map[rune][]*condensedField) {
	for _, tok := range HandleTokens {
		newField := &condensedField{
			Name:      field.Name,
			Type:      field.Type,
			RequestID: eField.NameByPriority(field, eField.PriorityJsonBson),
		}

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
