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
func WriteToField(field *reflect.Value, data interface{}) (err error) {
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
	}

	return nil
}

/*
CheckCC returns whether the given field's type is a
collection type (array, slice, ...) as well as the
type of an element in the collection.
*/
func CheckCC(field reflect.StructField) (bool, reflect.Type) {
	switch field.Type.Kind() {
	default:
		return false, nil
	case reflect.Slice, reflect.Array:
		return true, field.Type.Elem()
	}
}
