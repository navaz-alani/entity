package multiplexer

import (
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/navaz-alani/entity/entityErrors"
)

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

// duplicate ID tag 3
type ENoDupID3 struct {
	F1 int `json:"f_1" bson:"f1" _id_:"<id>"`
	// ID reset by field below
	F2 int `json:"f_2" bson:"f2" _id_:"<new_id>"`
}

// no database collection needed
type ENoDBColl struct {
	// entityID will still be "no-coll"
	F1 int `json:"f1" _id_:"!no-coll"`
}

// database type for mocking
type TestDB struct{}

// mock Collection function
func (db TestDB) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	return &mongo.Collection{}
}

func TestCreateDBUninitialized(t *testing.T) {
	_, err := Create(nil)
	if err != entityErrors.DBUninitialized {
		t.Fail()
	}
}

func TestCreateNoID(t *testing.T) {
	_, err := Create(db, ENoID{})
	if err == nil {
		t.Fail()
	}
}

func TestCreateDupID(t *testing.T) {
	_, err := Create(TestDB{}, EDupID1{}, EDupID2{})
	if err == nil {
		t.Fail()
	}
}

func TestCreateNoDupID(t *testing.T) {
	_, err := Create(TestDB{}, EDupID2{}, ENoDupID3{})
	if err != nil {
		t.Fail()
	}
}

func TestCreateNoCollection(t *testing.T) {
	mux, err := Create(TestDB{}, ENoDBColl{})
	if err != nil {
		t.Fatal(err)
	}

	if coll := mux.Collection("no-coll"); coll != nil {
		t.Fail()
	}
}
