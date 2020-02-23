package entityErrors

import "fmt"

var (
	IncompatibleEntityType = fmt.Errorf("incompatible entity type")
	UndefinedAxis          = fmt.Errorf("entity axis undefined (Axis Policy)")
	DBDecodeFail           = fmt.Errorf("failed to decode DB value")
	AddedIDParseFail       = fmt.Errorf("added entity but failed to parse addedID")
	BodyIncomplete         = fmt.Errorf("entity body incomplete; will not add")
)
