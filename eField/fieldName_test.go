package eField_test

import (
	"reflect"
	"testing"

	fName "github.com/navaz-alani/entity/eField"
)

type TestStruct struct {
	TSField1 string `json:"ts1_field" bson:"ts1Field"`
	TSField2 string `json:"ts2_field" bson:"-"`
}

func TestByPriorityAllTags(t *testing.T) {
	TSField1 := reflect.TypeOf(TestStruct{}).Field(0)

	if res := fName.NameByPriority(TSField1, fName.PriorityBsonJson); res != TSField1.Tag.Get("bson") {
		t.Fail()
	}
	if res := fName.NameByPriority(TSField1, fName.PriorityJsonBson); res != TSField1.Tag.Get("json") {
		t.Fail()
	}
	if res := fName.NameByPriority(TSField1, fName.Priority{Tags: []string{}}); res != TSField1.Name {
		t.Fail()
	}
}

func TestByPriorityMissingTags(t *testing.T) {
	TSField2 := reflect.TypeOf(TestStruct{}).Field(1)

	if res := fName.NameByPriority(TSField2, fName.PriorityBsonJson); res != TSField2.Tag.Get("json") {
		t.Fail()
	}
	if res := fName.NameByPriority(TSField2, fName.PriorityJsonBson); res != TSField2.Tag.Get("json") {
		t.Fail()
	}
	if res := fName.NameByPriority(TSField2, fName.Priority{Tags: []string{}}); res != TSField2.Name {
		t.Fail()
	}
}
