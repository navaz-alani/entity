package entityErrors

import "fmt"

var WriteToPtrField = fmt.Errorf("attempt to write to a pointer valued field")
