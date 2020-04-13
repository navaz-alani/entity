package entity_test

import (
	"testing"

	"github.com/navaz-alani/entity"
)

func TestNewEntity_NonStruct(t *testing.T) {
	nonStructDefs := []interface{}{
		"string",
		123,
		[]interface{}{},
	}

	for i := 0; i < len(nonStructDefs); i++ {
		if _, err := entity.NewEntity(nonStructDefs[i]); err == nil {
			t.Fatal("expected non-struct entity definition to error")
		}
	}
}

type UserNoID struct {
	Name  string `json:"name" .rq:"t"`
	Email string `json:"email" .ax:"t" .rq:"t"`
}

// IDTag is required for every definition
func TestNewEntity_NoID(t *testing.T) {
	_, err := entity.NewEntity(UserNoID{})
	if err == nil {
		t.Fatal("expected error for entity with no ID")
	}
}

type User struct {
	Name  string `.id:"user" .rq:"t"`
	Email string `.id:"id_ignored" .ax:"t" .rq:"t" .va:"rep/email/"`
}

const userID = "user"

var userEty *entity.Entity

func TestNewEntityParse(t *testing.T) {
	ety, err := entity.NewEntity(User{})
	if err != nil {
		t.Log(err)
		t.Fatal("^ got error when compiling user entity")
	}
	userEty = ety
}

func TestEntity_ID(t *testing.T) {
	if id := userEty.ID(); id != userID {
		t.Fatalf("invalid id, expected %s, got %s", userID, id)
	}
}

func TestEntity_TypeCheck(t *testing.T) {
	ety, err := entity.NewEntity(User{})
	if err != nil {
		t.Log(err)
		t.Fatal("^ got error when compiling user entity")
	}

	if ety.TypeCheck(UserNoID{}) {
		t.Fatal("expected mismatched type error")
	}

	if !ety.TypeCheck(User{}) {
		t.Fatal("unexpected type mismatch error")
	}
}

func TestEntity_Validate(t *testing.T) {
	testUsr := User{
		Name:  "Test user",
		Email: "invalid.@email_com",
	}
	// test invalid email is invalidated
	if err := userEty.Validate(testUsr); err == nil {
		t.Log("expected user to be invalid")
		t.Fail()
	}

	// test valid email passes
	testUsr.Email = "validEmail@test.com"
	if err := userEty.Validate(testUsr); err != nil {
		t.Log("expected no validation error")
		t.Fail()
	}
}
