package minigrush

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func do(request *http.Request, listener *Listener) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	listener.ServeHTTP(response, request)
	return response
}

func doRequest(request *http.Request, t *testing.T) (*httptest.ResponseRecorder, *Listener) {
	store := dummy.Store{}
	channel := make(chan *Petition, 1000)
	listener := &Listener{
		SendTo:        channel,
		PetitionStore: store,
	}
	response := do(request, listener)
	return response, listener
}
func doAny(listener *Listener, t *testing.T) *httptest.ResponseRecorder {
	request, err := http.NewRequest("GET", urls[0], bytes.NewReader(data[0]))
	request.Header.Set(RelayerHost, targetHosts[0])
	if err != nil {
		t.Fatal(err)
	}
	response := do(request, listener)
	return response
}

func TestMissingHeader(t *testing.T) {
	for _, method := range methods {
		request, err := http.NewRequest(method, urls[0], bytes.NewReader(data[0]))
		if err != nil {
			t.Fatal(err)
		}
		response, listener := doRequest(request, t)
		if response.Code != 400 {
			t.Errorf("missing x-relayer-host should return 400 %d", response.Code)
		}
		if len(listener.SendTo) > 0 {
			t.Errorf("invalid request should not be enqueued, method %q len(listener.SendTo) %d", method, len(listener.SendTo))
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
		response, l := doRequest(request, t)
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
func TestStop(t *testing.T) {
	listener := &Listener{
		SendTo:        make(chan *Petition, 1000),
		PetitionStore: dummy.Store{},
	}
	listener.Stop()
	response := doAny(listener, t)
	if response.Code != 503 {
		t.Errorf("Listener stopping should return 503 code, not %d", response.Code)
	}
}
func TestFullQueue(t *testing.T) {
	listener := &Listener{
		SendTo:        make(chan *Petition), //not buffered, would block with first petition, simulate full queue
		PetitionStore: dummy.Store{},
	}
	response := doAny(listener, t)
	if response.Code != 500 {
		t.Errorf("Listener with full queue should return 500 code, not %d", response.Code)
	}
}

type BadStore struct {
	dummy.Store
}

func (bs *BadStore) Put(id string, data []byte) error {
	return fmt.Errorf("error caused for testing bad Put")
}
func TestBadStore(t *testing.T) {
	listener := &Listener{
		SendTo:        make(chan *Petition, 10), //not buffered, would block with first petition, simulate full queue
		PetitionStore: &BadStore{},
	}
	response := doAny(listener, t)
	if response.Code != 500 {
		t.Errorf("Listener with full queue should return 500 code, not %d", response.Code)
	}
}
