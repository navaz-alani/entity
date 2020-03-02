package multiplexer

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/navaz-alani/entity/entityErrors"
	"github.com/navaz-alani/entity/multiplexer/muxContext"
)

/*
The following is a simple mechanism for running multiple tests on the
middleware generator to test for parsing JSON.
*/

/*
reqTest defines the structure of a multiplexer middleware
creation request parsing test.
*/
type reqTest struct {
	Definitions    []interface{}
	EntityID       string
	JSONPayload    string
	ExpectedEntity interface{}
}

/*
requestTests is an array of the reqTests to be carried out.
*/
var requestTests = [...]reqTest{
	{
		[]interface{}{TestUser{}},
		"user", DummyUserDataJSON,
		DummyUserData,
	},
	{
		[]interface{}{UserEmbed{}, Task{}, TaskDetails{}},
		"user-embed", dummyEmbedDataJSON,
		DummyUserEmbed,
	},
	{
		[]interface{}{EmbedCollUser{}, Task{}, TaskDetails{}},
		"user-embed-coll", dummyEmbedCollDataJSON, DummyEmbedCollUser,
	},
	{
		[]interface{}{Project{}, TestSuite{}, TestCase{}},
		"project", DummyProjectJSON,
		DummyProject,
	},
}

func TestEntityMux_CreationMiddlewareNoCHandleFields(t *testing.T) {
	mux, err := Create(TestDB{}, EDupID1{})
	if err != nil {
		t.Failed()
		return
	}

	_, err = mux.CreationMiddleware("<id>")
	if err != entityErrors.NoClassificationFields {
		t.Fail()
	}
}

func EntityMux_CreationMiddlewareRequestParseTestHelper(t *testing.T, rt *reqTest) {
	mux, err := Create(TestDB{}, rt.Definitions...)
	if err != nil {
		t.Fatal(err)
	}

	hd, err := mux.CreationMiddleware(rt.EntityID)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", bytes.NewReader([]byte(rt.JSONPayload)))
	if err != nil {
		t.Fatal(err)
	}

	verify := func(w http.ResponseWriter, r *http.Request) {
		muxCtx, err := muxContext.IsolateCtx(r)
		if err != nil {
			t.Fatal(err)
		}
		data := muxCtx.Retrieve(rt.EntityID)

		if !reflect.DeepEqual(data, rt.ExpectedEntity) {
			log.Print("got:      ", data)
			log.Print("expected: ", rt.ExpectedEntity)
			t.Fatal()
		}
	}

	handler := hd(http.HandlerFunc(verify))
	handler.ServeHTTP(httptest.NewRecorder(), req)
}

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Define tests
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func TestEntityMux_CreationMiddlewareRequestParse(t *testing.T) {
	EntityMux_CreationMiddlewareRequestParseTestHelper(t, &requestTests[0])
}

func TestEntityMux_CreationMiddlewareRequestParseEmbedded(t *testing.T) {
	EntityMux_CreationMiddlewareRequestParseTestHelper(t, &requestTests[1])
}

func TestEntityMux_CreationMiddlewareRequestCollectionEmbed(t *testing.T) {
	EntityMux_CreationMiddlewareRequestParseTestHelper(t, &requestTests[2])
}

func TestEntityMux_CreationMiddlewareRequestCollectionsEmbedDeep(t *testing.T) {
	EntityMux_CreationMiddlewareRequestParseTestHelper(t, &requestTests[3])
}
