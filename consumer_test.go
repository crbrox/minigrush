// consumer_test.go
package minigrush

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/crbrox/store/dummy"
)

const idPetition = "1"

var testBody = []byte("sent content to target host")
var testBodyResponse = []byte("Hello, client (from target host)\n")

func TestAll(t *testing.T) {
	storeP := dummy.Store{}
	storeR := dummy.Store{}
	channel := make(chan *Petition, 1000)
	consumer := &Consumer{GetFrom: channel, PetitionStore: storeP, ReplyStore: storeR}
	resultCh := make(chan *http.Request, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, string(testBodyResponse))
		rcvdBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(rcvdBody, testBody) {
			t.Error("received body in target host is not equal to sent body")
		}
		resultCh <- r
	}))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/x/y/z"
	u.RawQuery = "q=a&r=b"
	petition := &Petition{
		Id:           idPetition,
		TargetHost:   u.Host,
		TargetScheme: u.Scheme,
		Method:       "GET",
		URL:          u,
		Proto:        "HTTP/1.1",
		Body:         testBody,
		RemoteAddr:   "127.0.0.1",
		Host:         u.Host,
		Created:      time.Now(),
	}
	endCh := consumer.Start(2)
	channel <- petition
	rcvRequest := <-resultCh
	if rcvRequest.URL.Path != petition.URL.Path {
		t.Errorf("received url path is not equal to sent url path %q %q", rcvRequest.URL, petition.URL)
	}
	if rcvRequest.URL.RawQuery != petition.URL.RawQuery {
		t.Errorf("received query is not equal to sent query %q %q", rcvRequest.URL, petition.URL)
	}
	if rcvRequest.Method != petition.Method {
		t.Errorf("received method is not equal to sent method %q %q", rcvRequest.URL, petition.URL)
	}
	consumer.Stop()
	select {
	case <-endCh:
	case <-time.After(time.Second):
		t.Error("time out stopping")
	}
	br, err := storeR.Get(idPetition)
	if err != nil {
		t.Errorf("reply should be stored %v", err)
	}
	reply := &Reply{}
	err = json.Unmarshal(br, reply)
	if err != nil {
		t.Errorf("reply should be json-encoded %v", err)
	}
	if !reflect.DeepEqual(reply.Body, testBodyResponse) {
		t.Errorf("target response body does not match %v %v", reply.Body, testBodyResponse)
	}
	_, err = storeP.Get(idPetition)
	if err == nil {
		t.Error("petition should be deleted %q", idPetition)
	}
}
