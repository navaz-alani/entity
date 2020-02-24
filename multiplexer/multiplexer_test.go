package multiplexer_test

import (
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/navaz-alani/entity/entityErrors"
	"github.com/navaz-alani/entity/multiplexer"
)

var eMux *multiplexer.EntityMux
var db = &mongo.Database{}

// no ID tag
type ENoID struct {
	F1 int `json:"f_1" bson:"f1" _id_:""`
	F2 int `json:"f_2" bson:"f2" _id_:"-"`
}

// duplicate ID tag 1
type EDupID1 struct {
	F1 int `json:"f_1" bson:"f1" _id_:"<id>"`
}

// duplicate ID tag 2
type EDupID2 struct {
	F2 int `json:"f_2" bson:"f2" _id_:"<id>"`
}

// database type for mocking
type TestDB struct{}

// mock Collection function
func (db TestDB) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	return &mongo.Collection{}
}

func TestCreateDBUninitialized(t *testing.T) {
	_, err := multiplexer.Create(nil)
	if err != entityErrors.DBUninitialized {
		t.Fail()
	}
}

func TestCreateNoID(t *testing.T) {
	_, err := multiplexer.Create(db, ENoID{})
	if err == nil {
		t.Fail()
	}
}

func TestCreateDupID(t *testing.T) {
	_, err := multiplexer.Create(TestDB{}, EDupID1{}, EDupID2{})
	if err == nil {
		t.Fail()
	}
}
