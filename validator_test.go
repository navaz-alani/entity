package entity_test

import (
	"testing"

	"github.com/navaz-alani/entity"
)

var ValidEmails = []string{
	"email@example.com",
	"firstname.lastname@example.com",
	"email@subdomain.example.com",
	"email@subsubdomain.subdomain.example.com",
	"firstname+lastname@example.com",
}

var InvalidEmails = []string{
	"email.example.com",
	"email@example@example.com",
	".email@example.com",
	"email.@example.com",
	"email..email@example.com",
	"",
}

func TestStringValidator_Email(t *testing.T) {
	emailValidator := entity.StringValidator(`rep/email/`)
	for _, email := range ValidEmails {
		if err := emailValidator.Validate(email); err != nil {
			t.Logf("%s invalidated but valid", email)
			t.Fail()
		}
	}
	for _, email := range InvalidEmails {
		if err := emailValidator.Validate(email); err == nil {
			t.Logf("%s validated but invalid", email)
			t.Fail()
		}
	}
}
