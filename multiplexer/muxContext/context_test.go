package muxContext

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/navaz-alani/entity/entityErrors"
)

type TestUser struct {
	ID    primitive.ObjectID `_id_:"user" json:"-" bson:"_id"`
	Name  string             `json:"name" _hd_:"c"`
	Email string             `json:"email" axis:"true" index:"text" _hd_:"c"`
}

type TestData struct{}

func (td TestData) Read(p []byte) (n int, err error) {
	return 1, nil
}

var (
	mux *EMuxContext

	keyStr = "<keyStr>"
	valStr = "<payload_data>"

	keyStruct = "<keyStruct>"
	valStruct = &TestUser{
		Name:  "Test User",
		Email: "test@test.com",
	}
)

func TestCreate(t *testing.T) {
	if res := Create(); res == nil {
		t.Error("eMux init fail; cannot proceed")
	} else {
		mux = res
	}
}

func TestEMuxContext_SetStr(t *testing.T) {
	mux.Set(keyStr, valStr)

	if mux.payloads[keyStr] != valStr {
		t.Fail()
	}
}

func TestEMuxContext_RetrieveStr(t *testing.T) {
	if res := mux.Retrieve(keyStr); res != valStr {
		t.Fail()
	}
}

func TestEMuxContext_SetStructPtr(t *testing.T) {
	mux.Set(keyStruct, valStruct)

	if mux.payloads[keyStruct] != valStruct {
		t.Fail()
	}
}

func TestEMuxContext_RetrieveStructPtr(t *testing.T) {
	if res := mux.Retrieve(keyStruct); res != valStruct {
		t.Fail()
	} else {
		usr, _ := res.(*TestUser)

		if !reflect.DeepEqual(*usr, *valStruct) {
			t.Fail()
		}
	}
}

func TestIsolateCtxNoCtxInReq(t *testing.T) {
	req, _ := http.NewRequest("GET", "test.com", TestData{})

	if _, err := IsolateCtx(req); err != entityErrors.MuxCtxNotFound {
		t.Fail()
	}
}

func TestIsolateCtxCorruptCtx(t *testing.T) {
	req, _ := http.NewRequest("GET", "test.com", TestData{})
	reqWithCtx := req.WithContext(context.WithValue(context.Background(), muxCtxKey, ""))

	if _, err := IsolateCtx(reqWithCtx); err != entityErrors.MuxCtxCorrupt {
		t.Fail()
	}
}

func TestEMuxContext_EmbedCtx(t *testing.T) {
	req, _ := http.NewRequest("GET", "test.com", TestData{})
	reqWithCtx := mux.EmbedCtx(req, context.TODO())

	mux, _ := IsolateCtx(reqWithCtx)
	usr := mux.Retrieve(keyStruct).(*TestUser)

	if usr == nil || !reflect.DeepEqual(*usr, *valStruct) {
		t.Fail()
	}
}
