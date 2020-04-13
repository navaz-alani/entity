package entity

import "fmt"

type Err struct {
	msg  string
	Type string
}

func (e Err) Error() string { return e.msg }

const (
	TypeMismatch     = "type-mismatch"
	TagUndefined     = "tag-undefined"
	NonStruct        = "non-struct-kind"
	IncompleteValue  = "val-incomplete"
	InputInvalid     = "input-invalid"
	InputTypeInvalid = "input-type-invalid"
)

var (
	ErrMismatchedTypes  = Err{"mismatched types", TypeMismatch}
	ErrNonStruct        = Err{"non-struct definition", NonStruct}
	ErrInputTypeInvalid = Err{"unexpected input type", InputTypeInvalid}
	ErrInputInvalid     = Err{"input validation fail", InputInvalid}
)

func ExpectedTagDef(tag string) error {
	return Err{
		msg:  fmt.Sprintf("undefined tag '%s'", tag),
		Type: TagUndefined,
	}
}
