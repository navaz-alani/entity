package entity

import (
	"fmt"
	"reflect"
)

type Tag struct {
	Tag      string
	Required bool
}

type TagValue struct {
	// Val is the value of the tag in the entity definition.
	Val string
	// Provider is the index of the field which provides
	// the above value.
	Provider int
}

var (
	IDTag       = Tag{".id", true}
	IndexTag    = Tag{".ix", false}
	ValidateTag = Tag{".va", false}
	Tags        = []Tag{
		IDTag,
		IndexTag,
		ValidateTag,
	}
)

// Entity is a parsed version of a struct definition.
type Entity struct {
	Definition reflect.Type
	Specs      map[Tag][]TagValue
	Validation map[int]Validator
}

// ID returns the Entity's identifier, specified by the first
// value of the IDTag in the definition.
func (ety *Entity) ID() string { return ety.Specs[IDTag][0].Val }

// NewEntity creates a new Entity from the given struct definition.
func NewEntity(e interface{}) (*Entity, error) {
	eDef := reflect.TypeOf(e)
	// verify that e is indeed a struct
	if kind := eDef.Kind(); kind != reflect.Struct {
		return nil, ErrNonStruct
	}
	return parseDefinition(eDef)
}

// parseDefinition uses the field tags in the definition provided to
// compile a profile of the entity.
func parseDefinition(def reflect.Type) (ety *Entity, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(r.(string))
		}
	}()
	defTags := make(map[Tag][]TagValue)

	fieldCount := def.NumField()
	// for each tag, iterate over fields to get value of tag
	for _, tag := range Tags {
		for i := 0; i < fieldCount; i++ {
			field := def.Field(i)
			if val := field.Tag.Get(tag.Tag); val != "" {
				defTags[tag] = append(defTags[tag], TagValue{
					Val:      val,
					Provider: i,
				})
			}
		}
		if tag.Required && defTags[tag] == nil {
			return nil, ExpectedTagDef(tag.Tag)
		}
	}

	compiledEty := &Entity{
		Definition: def,
		Specs:      defTags,
		Validation: make(map[int]Validator),
	}

	validationTags := defTags[ValidateTag]
	for _, tag := range validationTags {
		fieldType := def.Field(tag.Provider).Type.Kind()
		switch fieldType {
		case reflect.String:
			compiledEty.Validation[tag.Provider] = StringValidator(tag.Val)
		default:
			panic(fmt.Sprintf("validation for type '%s' not supported", fieldType))
		}
	}

	return compiledEty, nil
}

func (ety *Entity) TypeCheck(e interface{}) bool {
	return reflect.TypeOf(e) == ety.Definition
}

// Create creates an instance of the entity in the database.
func (ety *Entity) Create(e interface{}) error {
	if !ety.TypeCheck(e) {
		return ErrMismatchedTypes
	}

	return nil
}

// Validate tests the fields of e which, in the Entity definition,
// possessed the validation tag against the Validator generated
// from that tag. If any of the fields fails, ErrValidation is
// returned, otherwise the returned error should be nil.
func (ety *Entity) Validate(e interface{}) error {
	inputVal := reflect.ValueOf(e)
	for fIndex, validator := range ety.Validation {
		fieldToValidate := inputVal.Field(fIndex)
		if err := validator.Validate(fieldToValidate.Interface()); err != nil {
			return err
		}
	}
	return nil
}
