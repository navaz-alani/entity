package fieldName

import (
	"reflect"
)

const (
	JSONTag string = "json"
	BSONTag string = "bson"
)

/*
Priority is a type used to define the order of
preference of available field name options.
*/
type Priority struct {
	Tags []string
}

var (
	// Choose first of BSON tag, JSON tag, Field name
	PriorityJsonBson = Priority{Tags: []string{JSONTag, BSONTag}}
	// Choose first of JSON tag, BSON tag, Field name
	PriorityBsonJson = Priority{Tags: []string{BSONTag, JSONTag}}
)

/*
ByPriority returns the name of the field using the priority
p given.
When the tags in p.Tags have been exhausted, the field's name
is returned. Therefore this function is guaranteed to return a
name for the field.
*/
func ByPriority(field reflect.StructField, p Priority) string {
	for _, tagName := range p.Tags {
		if tag := field.Tag.Get(tagName); tag != "" && tag != "-" {
			return tag
		}
	}
	return field.Name
}
