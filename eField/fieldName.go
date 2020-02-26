/*
Package eField provides definitions and functions which
are used during internal operations on reflect.StructField
types.
*/
package eField

import (
	"reflect"
)

/*
These are tags which are related to the field's name.
*/
const (
	JSONTag string = "json"
	BSONTag string = "bson"
)

/*
Priority is a type used to define the order of
preference of available eField name options.
*/
type Priority struct {
	Tags []string
}

/*
These are priorities for choosing a field name.
*/
var (
	// Choose first of BSON tag, JSON tag, Field name
	PriorityJsonBson = Priority{Tags: []string{JSONTag, BSONTag}}
	// Choose first of JSON tag, BSON tag, Field name
	PriorityBsonJson = Priority{Tags: []string{BSONTag, JSONTag}}
)

/*
NameByPriority returns the name of the eField using the priority
p given.
When the tags in p.Tags have been exhausted, the eField's name
is returned. Therefore this function is guaranteed to return a
name for the eField.
*/
func NameByPriority(field reflect.StructField, p Priority) string {
	for _, tagName := range p.Tags {
		if tag := field.Tag.Get(tagName); tag != "" && tag != "-" {
			return tag
		}
	}
	return field.Name
}
