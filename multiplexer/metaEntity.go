package multiplexer

import (
	"reflect"
	"strings"

	"github.com/navaz-alani/entity"
)

/*
condensedField is a shorthand representation of the information
which is commonly used/is important within the context of this
package.
*/
type condensedField struct {
	/*
		Name is the field's field name.
	*/
	Name string
	/*
		Type is the reflection type of the field.
	*/
	Type reflect.Type
	/*
		RequestID is a string which specifies the fieldName to
		expect when parsing JSON for this field.
		It can be equal to the Name field.
	*/
	RequestID string
	/*
		StructIndex is an integer array which describes the
		embedded index of a field in a struct.
	*/
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
   The following constants define the tokens for field
   classifications in a metaEntity. These tokens are to be
   used within the context of the entity.HandleTag.
*/
const (
	/*
		CollectionNameToken maps to an array containing a
		single (or none at all) pointer to a condensedField
		whose RequestID is the Entity's collectionName in the
		database.
	*/
	CollectionNameToken rune = '*'
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
requestID returns the fieldName to expect when parsing for this
field in incoming request JSON payloads.

TODO: a general requestID with priority function
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

TODO: generalize this function for all tags
*/
func classifyFields(defType reflect.Type) map[rune][]*condensedField {
	classifications := map[rune][]*condensedField{}

	collectionName := classifications[CollectionNameToken]
	creationFields := classifications[CreationFieldsToken]

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
