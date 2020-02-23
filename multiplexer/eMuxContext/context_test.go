package eMuxContext_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"

	muxCtx "github.com/navaz-alani/entity/multiplexer/eMuxContext"
)

type TestUser struct {
	ID    primitive.ObjectID `_id_:"user" json:"-" bson:"_id"`
	Name  string             `json:"name" _hd_:"c"`
	Email string             `json:"email" axis:"true" index:"text" _hd_:"c"`
}

type TestReqData struct {}
func (td TestReqData) Read(p []byte)(n int, err error) {
	return 1, nil
}

var (
	mux *muxCtx.EMuxContext

	keyStr = "<keyStr>"
	valStr = "<payload_data>"

	keyStruct = "<keyStruct>"
	valStruct = &TestUser{
		Name:  "Test User",
		Email: "test@test.com",
	}
)

func TestCreate(t *testing.T) {
	if res := muxCtx.Create(); res == nil {
		t.Error("eMux init fail; cannot proceed")
	} else {
		mux = res
	}
}

func TestEMuxContext_SetStr(t *testing.T) {
	mux.Set(keyStr, valStr)

	if mux.Payloads[keyStr] != valStr {
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

	if mux.Payloads[keyStruct] != valStruct {
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

func  TestEMuxContext_ContextualizeRequest(t *testing.T) {
	req, _ := http.NewRequest("GET","test.com",TestReqData{})
	reqWithCtx := mux.ContextualizeRequest(req,context.TODO(), muxCtx.EMuxKey)

	mux, _ := reqWithCtx.Context().Value(muxCtx.EMuxKey).(*muxCtx.EMuxContext)
	usr := mux.Retrieve(keyStruct).(*TestUser)

	if usr == nil || !reflect.DeepEqual(*usr, *valStruct) {
		t.Fail()
	}
}
