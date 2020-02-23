package multiplexer

import (
	"reflect"
	"strings"

	"github.com/navaz-alani/entity"
	"github.com/navaz-alani/entity/fieldName"
)

/*
condensedField is a shorthand representation of the information
which is commonly used/is important within the context of this
package.
*/
type condensedField struct {
	// Name is the field's field name.
	Name string
	// Type is the reflection type of the field.
	Type reflect.Type
	/*
		RequestID is a string which specifies the fieldName to
		expect when parsing JSON for this field.
		It can be equal to the Name field.
	*/
	RequestID string
	Value     interface{}
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
		a slice of pointers to condensedFields.
	*/
	FieldClassifications map[rune][]*condensedField
}

/*
   The following constants define the tokens for field
   classifications in a metaEntity. These tokens are to be
   used within the context of the entity.HandleTag.
*/
const (
	/*
		CollectionIDToken maps to an array containing a
		single (or none at all) pointer to a condensedField
		whose RequestID is the Entity's collectionName in the
		database.
	*/
	CollectionIDToken rune = '*'
	/*
		CreationFieldsToken maps to an array containing fields
		which were specified to be provided in an http.Request.
	*/
	CreationFieldsToken rune = 'c'
	/*
		AxisFieldToken is maps to an array containing fields which
		are tagged as axis fields.
	*/
	AxisFieldToken rune = 'a'
)

/*
HandleTokens defines the set of tokens which can be used in
the entity.HandleTag of a struct field for classification.
*/
var HandleTokens = []rune{
	CreationFieldsToken,
	AxisFieldToken,
}

/*
classifyFields is a function which iterates over the fields of
the given Type and classifies them by their HandleTags.
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
	for _, tok := range HandleTokens {
		newField := &condensedField{
			Name:      field.Name,
			Type:      nil,
			RequestID: fieldName.ByPriority(field, fieldName.PriorityJsonBson),
		}

		if classes[tok] == nil {
			classes[tok] = make([]*condensedField, 0)
		}

		if tag := field.Tag.Get(entity.IDTag); tag != "" && tag != "-" {
			newField.Value = tag
			classes[CollectionIDToken] = make([]*condensedField, 0)
			classes[CollectionIDToken] = append(classes[CollectionIDToken], newField)
		}

		if tag := field.Tag.Get(entity.HandleTag); strings.ContainsAny(tag, string(tok)) {
			classes[tok] = append(classes[tok], newField)
		}
	}
}
