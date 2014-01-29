// recoverer_test.go
package minigrush

import (
	"encoding/json"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/crbrox/store/dummy"
)

func TestA(t *testing.T) {
	url1, _ := url.Parse("http://golang.org/pkg/net/http/#NewRequest")
	url2, _ := url.Parse("https://www.google.es")
	p1 := &Petition{
		Id:           "id1",
		Body:         []byte("body1"),
		Method:       "GET",
		URL:          url1,
		Proto:        "HTTP/1.1",
		Header:       make(http.Header),
		Trailer:      make(http.Header),
		RemoteAddr:   "RemoteAddr1",
		RequestURI:   "RequestURI1",
		TargetHost:   "targetHost1",
		TargetScheme: "http",
		Created:      time.Now()}
	p2 := &Petition{
		Id:           "id2",
		Body:         []byte("body2"),
		Method:       "POST",
		URL:          url2,
		Proto:        "HTTP/1.1",
		Header:       make(http.Header),
		Trailer:      make(http.Header),
		RemoteAddr:   "RemoteAddr2",
		RequestURI:   "RequestURI2",
		TargetHost:   "targetHost2",
		TargetScheme: "https",
		Created:      time.Now()}
	var petitions = []*Petition{p1, p2}
	petStore := dummy.Store{}
	petCh := make(chan *Petition, len(petitions))
	for _, p := range petitions {
		text, err := json.Marshal(p)
		if err != nil {
			t.Error(err)
		}
		petStore.Put(p.Id, text)
	}
	recoverer := &Recoverer{
		SendTo:        petCh,
		PetitionStore: petStore,
	}
	err := recoverer.Recover()
	if err != nil {
		t.Error(err)
	}
	if len(petCh) < len(petitions) {
		t.Fatalf("all stored petitions should be enqueued len(petCh) %d len(pseudoPetitions) %d ",
			len(petCh), len(petitions))
	}
	for i := 0; i < len(petitions); i++ {
		select {
		case ep := <-petCh:
			var sp *Petition
			switch ep.Id {
			case "id1":
				sp = p1
			case "id2":
				sp = p2
			default:
				t.Fatalf("unknown enqueued petition id %q", ep.Id)
			}
			if !reflect.DeepEqual(ep, sp) {
				t.Fatal("enqued petition is not equal to stored petition")
			}
		default:
			t.Fatal("should be more enqueued petitions")
		}

	}
}
