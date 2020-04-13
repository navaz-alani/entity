package entity

import (
	"fmt"
	"regexp"
)

// Validator describes a type which verifies the value of a
// field in an entity.
type Validator interface {
	Validate(interface{}) error
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// STRING VALIDATOR
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// StrValidator is a type which validates inputs to fields
// taking string inputs. Validation is done primarily by
// using regular expressions.
type StrValidator struct {
	re *regexp.Regexp
}

const emailRegexp = "^(([^<>()\\[\\]\\\\.,;:\\s@\"]+(\\.[^<>()\\[\\]\\\\.," +
	";:\\s@\"]+)*)|(\".+\"))@((\\[[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]" +
	"{1,3}\\])|(([a-zA-Z\\-0-9]+\\.)+[a-zA-Z]{2,}))$"

var (
	// PresetNone accepts anything.
	PresetNone = regexp.MustCompile(`.*`)
	// PresetEmail checks the input against the following email
	// regular expression pattern.
	PresetEmail = regexp.MustCompile(emailRegexp)
)

func (sv StrValidator) Validate(input interface{}) error {
	if strInput, ok := input.(string); !ok {
		return ErrInputTypeInvalid
	} else {
		if sv.re.MatchString(strInput) {
			return nil
		}
		return ErrInputInvalid
	}
}

// StringValidator parses the validation tag of a field whose type is
// a string and returns a validator for that field.
// A regular expression can be specified in the validation tag in one
// of two formats.
//
// A raw regular expression, which is just a plain old regexp. This
// is specified with the format re/<regexp>/, where <regexp> is the
// regular expression to be used to validate the input.
//
// A preset regular expression is a string specifying the type of string
// that the field should accept. An example is an email or phone number.
// Presets can be specified using the format rep/<preset_name>/, where
// <preset_name> is one of "email".
func StringValidator(tag string) Validator {
	var validator StrValidator
	reRaw := regexp.MustCompile(`re/.*/`)
	rePreset := regexp.MustCompile(`rep/.*/`)

	// extract validation regexp from tag
	if match := reRaw.FindString(tag); match != "" {
		validator.re = regexp.MustCompile(match[3 : len(match)-1])
	} else if match := rePreset.FindString(tag); match != "" {
		var validationPattern *regexp.Regexp
		presetVal := match[4 : len(match)-1]
		switch presetVal {
		case "email":
			validationPattern = PresetEmail
		default:
			// no way to know how to validate the field
			// - could be a typo?
			panic(fmt.Sprintf("unknown validation preset '%s'", presetVal))
		}
		validator.re = validationPattern
	} else {
		validator.re = PresetNone
	}
	return validator
}
