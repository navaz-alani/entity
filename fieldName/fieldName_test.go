package fieldName

import (
	"reflect"
	"testing"
)

type TestStruct struct {
	TSField1 string `json:"ts1_field" bson:"ts1Field"`
	TSField2 string `json:"ts2_field" bson:"-"`
}

func TestByPriorityAllTags(t *testing.T) {
	TSField1 := reflect.TypeOf(TestStruct{}).Field(0)

	if res := ByPriority(TSField1, PriorityBsonJson); res != TSField1.Tag.Get("bson") {
		t.Fail()
	}
	if res := ByPriority(TSField1, PriorityJsonBson); res != TSField1.Tag.Get("json") {
		t.Fail()
	}
	if res := ByPriority(TSField1, Priority{Tags: []string{}}); res != TSField1.Name {
		t.Fail()
	}
}

func TestByPriorityMissingTags(t *testing.T) {
	TSField2 := reflect.TypeOf(TestStruct{}).Field(1)

	if res := ByPriority(TSField2, PriorityBsonJson); res != TSField2.Tag.Get("json") {
		t.Fail()
	}
	if res := ByPriority(TSField2, PriorityJsonBson); res != TSField2.Tag.Get("json") {
		t.Fail()
	}
	if res := ByPriority(TSField2, Priority{Tags: []string{}}); res != TSField2.Name {
		t.Fail()
	}
}
