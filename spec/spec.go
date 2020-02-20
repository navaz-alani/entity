/*
Package spec defines the ESpec (Entity Specification) Type,
which can be used for queries, and update operations on
the general Entity.
*/
package spec

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

/*
ESpec is a format which can be used to specify
fields and values for entities. This is useful when
constructing queries, as well as when specifying
changes.

Ensure that Specs intended for update operations
have the zero value for the QueryOperator field
and vice versa.
*/
type ESpec struct {
	/*
		Field is the database field-name to constrain
		the search
	*/
	Field string `json:"field"`
	/*
		Target specifies the value to query in the
		constrained field
	*/
	Target interface{} `json:"target"`
	/*
		UpdateOperator specifies an update operator for
		this ESpec
	*/
	UpdateOperator string `json:"updateOperator"`
	/*
		QueryOperator specified a query operator for
		this ESpec
	*/
	QueryOperator string `json:"queryOperator"`
}

/*
ToBSON encodes the ESpec as BSON map which can be
used as a query filter.
For now, only use this with MongoDB comparison
operators as they have a consistent syntax.
*/
func (s *ESpec) ToBSON() bson.M {
	if s.QueryOperator == "" {
		return bson.M{s.Field: s.Target}
	}
	return bson.M{
		s.Field: bson.M{
			fmt.Sprintf("$%s", s.QueryOperator): s.Target,
		},
	}
}

/*
ToUpdateSpec returns a BSON map which can be used
as an update document. The ESpec's Operator field
must be a valid Mongo update operator.

The following update operators are used in the context
of other operators and are not supported:
	"$(update)", "$[]", "$[<identifier>]",
	"$slice", "$sort", "$each", "$position"
*/
func (s *ESpec) ToUpdateSpec() bson.M {
	// TODO: an optional parameter to allow contextual update operators
	operator := "set"
	if s.UpdateOperator != "" {
		operator = s.UpdateOperator
	}

	return bson.M{
		fmt.Sprintf("$%s", operator): s.ToBSON(),
	}
}
