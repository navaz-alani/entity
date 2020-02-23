package entityErrors

import "fmt"

var (
	/*
		IncompatibleEntityType is an error that signifies that the
		Entity provided for a certain operation cannot be used for
		that operation due to a type mismatch.
	*/
	IncompatibleEntityType = fmt.Errorf("incompatible entity type")
	/*
		UndefinedAxis is an error which signifies that an Entity's
		axis fields are undefined. This could be raised, for example,
		when attempting to Filter for an Entity whose ID field is
		primitive.NilObjectID and all axis fields are zeroed- in such
		a case, no Filter may be created.
	*/
	UndefinedAxis = fmt.Errorf("entity axis undefined (Axis Policy)")
	/*
		DBDecodeFail is an error which represents a failed attempt to
		read a value returned from the database.
	*/
	DBDecodeFail = fmt.Errorf("failed to decode DB value")
	/*
		AddedIDParseFail is an error which signifies that an operation
		to add an Entity has succeeded, but the attempt to read the
		new database ID of that Entity has failed.
	*/
	AddedIDParseFail = fmt.Errorf("added entity but failed to parse addedID")
	/*
		BodyIncomplete is an error which signifies that the required
		fields for an Entity have not been provided. It is returned
		when attempting to add an incomplete Entity to the database.
	*/
	BodyIncomplete = fmt.Errorf("entity body incomplete- will not add")
)
