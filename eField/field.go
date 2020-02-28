package eField

import (
	"reflect"

	"github.com/navaz-alani/entity/entityErrors"
)

/*
WriteToField takes a eField value and attempts to set
its value to the given data. If the given data cannot
successfully be assigned to the given field, a
entityErrors.InvalidDataType error is returned.

This function will NEVER write to a eField which stores
a pointer kind.
*/
func WriteToField(field reflect.Value, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = entityErrors.InvalidDataType
		}
	}()

	/*
		Do not need to support pointers because an Entity has database handles.
		Pointers stored in databases would make no sense and therefore there is
		no pointer case in this switch.
	*/
	switch field.Kind() {
	default:
		field.Set(reflect.ValueOf(data))
	case reflect.String:
		field.SetString(data.(string))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(data.(int64))
	case reflect.Float32, reflect.Float64:
		field.SetFloat(data.(float64))
	case reflect.Bool:
		field.SetBool(data.(bool))
	case reflect.Ptr:
		return entityErrors.WriteToPtrField
	}

	return nil
}

/*
CheckEmbedding returns values indicating the field's type embedding.
If the field is of collection kind (slice), the cFlag is set to
true and the embeddedType is set to the type of a single element
in that collection.
If the field is of struct kind, the sFlag is set to true and the
embeddedType is set to the field's type.
*/
func CheckEmbedding(field reflect.StructField) (cFlag, sFlag bool, embeddedType reflect.Type) {
	switch field.Type.Kind() {
	case reflect.Slice:
		cFlag = true
		embeddedType = field.Type.Elem()
	case reflect.Struct:
		sFlag = true
		embeddedType = field.Type
	}

	return cFlag, sFlag, embeddedType
}
