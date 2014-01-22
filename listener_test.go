package minigrush

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/crbrox/store/dummy"
)

var methods = []string{"GET", "POST", "PUT", "DELETE", "HEAD"}
var data = [][]byte{{1, 2, 3, 4, 0, 5, 6, 7}}
var urls = []string{"http://0.0.0.0:0"}
var targetHosts = []string{"0.0.0.0:0"}

func do(request *http.Request) (*httptest.ResponseRecorder, *Listener) {
	store := dummy.Store{}
	channel := make(chan *Petition, 1000)
	response := httptest.NewRecorder()
	l := &Listener{
		SendTo:        channel,
		PetitionStore: store,
	}
	l.ServeHTTP(response, request)
	return response, l
}

func TestMissingHeader(t *testing.T) {
	for _, method := range methods {
		request, err := http.NewRequest(method, urls[0], bytes.NewReader(data[0]))
		if err != nil {
			t.Fatal(err)
		}
		response, l := do(request)
		if response.Code != 400 {
			t.Errorf("missing x-relayer-host should return 400 %d", response.Code)
		}
		if len(l.SendTo) > 0 {
			t.Errorf("invalid request should not be enqueued, method %q len(l.SendTo) %d", method, len(l.SendTo))
		}
	}
}
func TestOK(t *testing.T) {
	for _, method := range methods {
		request, err := http.NewRequest(method, urls[0], bytes.NewReader(data[0]))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set(RelayerHost, targetHosts[0])
		response, l := do(request)
		if response.Code != 200 {
			t.Errorf("expected status code 200: %d", response.Code)
		}
		if len(l.SendTo) != 1 {
			t.Errorf("valid request should be enqueued, method %q len(l.SendTo) %d", method, len(l.SendTo))
		}
		id := strings.TrimSpace(response.Body.String())
		d, err := l.PetitionStore.Get(id)
		if err != nil {
			t.Errorf("petition should be stored with returned ID, method %q id %q err %v", method, id, err)
		}
		pet := &Petition{}
		err = json.Unmarshal(d, pet)
		if err != nil {
			t.Errorf("stored petition should be valid json, method %q id %q err %v", method, id, err)
		}
		if pet.Method != method {
			t.Errorf("petition's method should be equal to request's one, id %q method %q petition's method %q", id, method, pet.Method)
		}
		if !reflect.DeepEqual(pet.Body, data[0]) {
			t.Errorf("petition's body should be equal to request's one, id %q method %q", id, method)
		}
		if pet.TargetHost != targetHosts[0] {
			t.Errorf("petition's target host should be equal to target host, id %q method %q petition's host %q target host %q", id, method, pet.TargetHost, targetHosts[0])
			t.Error(string(d))
		}

	}
}
