package minigrush

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/crbrox/store/dummy"
)

func TestOk(t *testing.T) {
	const data = `
	{ a: 2, b: "b",
	guay : true,

	DOLLARS: 1234.5678
	}`
	const id = "1234_abc"
	replyer := &Replyer{
		ReplyStore: dummy.Store{},
	}
	replyer.ReplyStore.Put(id, []byte(data))
	request, err := http.NewRequest("GET", "/one/two/three/"+id, nil)
	if err != nil {
		t.Error(err)
	}
	response := httptest.NewRecorder()
	replyer.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Error("response should be 200 for existent reply")
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Error("content type of response should be application/json")
	}
	if !reflect.DeepEqual(response.Body.Bytes(), []byte(data)) {
		t.Error("response body should be equal to body reply")
	}
}
func TestNotFound(t *testing.T) {
	const id = "1234_abc"
	replyer := &Replyer{
		ReplyStore: dummy.Store{},
	}
	request, err := http.NewRequest("GET", "/one/two/three/"+id, nil)
	if err != nil {
		t.Error(err)
	}
	response := httptest.NewRecorder()
	replyer.ServeHTTP(response, request)
	if response.Code != 404 {
		t.Error("response should be 404 for inexistent reply")
	}
}
