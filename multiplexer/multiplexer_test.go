package multiplexer_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/navaz-alani/entity/entityErrors"
	"github.com/navaz-alani/entity/multiplexer"
	"github.com/navaz-alani/entity/multiplexer/muxContext"
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

// database type for mocking
type TestDB struct{}

// mock Collection function
func (db TestDB) Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection {
	return &mongo.Collection{}
}

// TestUser for middleware test
type TestUser struct {
	ID    primitive.ObjectID `json:"-" bson:"_id" _id_:"test-user"`
	Name  string             `json:"name" _hd_:"c"`
	Email string             `json:"email" _hd_:"c"`
	//Age   int64              `json:"age" _hd_:"c"`
}

var DummyUserData = TestUser{
	Name:  "Dummy User",
	Email: "dummy@user.com",
	//Age:   420,
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

func TestCreateNoDupID(t *testing.T) {
	_, err := multiplexer.Create(TestDB{}, EDupID2{}, ENoDupID3{})
	if err != nil {
		t.Fail()
	}
}

func TestEntityMux_CreationMiddlewareNoCHandleFields(t *testing.T) {
	mux, err := multiplexer.Create(TestDB{}, EDupID1{})
	if err != nil {
		t.Failed()
		return
	}

	_, err = mux.CreationMiddleware("<id>")
	if err != entityErrors.NoClassificationFields {
		t.Fail()
	}
}

func TestEntityMux_CreationMiddlewareRequestParse(t *testing.T) {
	mux, err := multiplexer.Create(TestDB{}, TestUser{})
	if err != nil {
		t.Failed()
		return
	}

	hd, err := mux.CreationMiddleware("test-user")
	if err != nil {
		t.Failed()
		return
	}

	payloadData, err := json.Marshal(DummyUserData)
	req, err := http.NewRequest("GET", "/", bytes.NewReader(payloadData))
	if err != nil {
		t.Fatal(err)
	}

	verify := func(w http.ResponseWriter, r *http.Request) {
		muxCtx, err := muxContext.IsolateCtx(r)
		if err != nil {
			t.Fatal()
		}

		data, ok := muxCtx.Retrieve("test-user").(*TestUser)
		if !ok {
			t.Fatal()
		}

		if !reflect.DeepEqual(*data, DummyUserData) {
			log.Print("got:      ", *data)
			log.Print("expected: ", DummyUserData)
			t.Fatal()
		}
	}

	handler := hd(http.HandlerFunc(verify))
	handler.ServeHTTP(httptest.NewRecorder(), req)
}
