package spec

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

var (
	querySpec1 = ESpec{
		Field:  "qs1-field",
		Target: "qs1",
	}

	querySpec2 = ESpec{
		Field:         "qs2-field",
		Target:        "qs2",
		QueryOperator: "in",
	}
)

func TestESpec_ToBsonNoQueryOp(t *testing.T) {
	expected := bson.M{"qs1-field": "qs1"}
	res := querySpec1.ToBSON()

	if !reflect.DeepEqual(expected, res) {
		t.Fail()
	}
}

func TestESpec_ToBsonWithQueryOp(t *testing.T) {
	expected := bson.M{"qs2-field": bson.M{"$in": "qs2"}}
	res := querySpec2.ToBSON()

	if !reflect.DeepEqual(expected, res) {
		t.Fail()
	}
}

var (
	updateSpec1 = ESpec{
		Field:  "us1-field",
		Target: "us1",
	}

	updateSpec2 = ESpec{
		Field:          "us2-field",
		Target:         "us2",
		UpdateOperator: "push",
	}
)

func TestESpec_ToUpdateSpecNoUpdateOp(t *testing.T) {
	expected := bson.M{"$set": bson.M{"us1-field": "us1"}}
	res := updateSpec1.ToUpdateSpec()

	if !reflect.DeepEqual(expected, res) {
		t.Fail()
	}
}

func TestESpec_ToUpdateSpecWithUpdateOp(t *testing.T) {
	expected := bson.M{"$push": bson.M{"us2-field": "us2"}}
	res := updateSpec2.ToUpdateSpec()

	if !reflect.DeepEqual(expected, res) {
		t.Fail()
	}
}
