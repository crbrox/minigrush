package grass

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/crbrox/store"
)

type Listener struct {
	SendTo        chan<- *Petition
	PetitionStore store.Interface
	Stopping      bool
	closeOnce     sync.Once
}

func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if l.Stopping {
		http.Error(w, "Estoy parando", 503)
		return
	}
	relayedRequest, e := newPetition(r)
	if e != nil {
		http.Error(w, e.Error(), 400)
		return
	}
	text, e := json.MarshalIndent(relayedRequest, "", " ")
	e = l.PetitionStore.Put(relayedRequest.Id, text)
	if e != nil {
		http.Error(w, relayedRequest.Id, 500)
		log.Println(relayedRequest.Id, e.Error())
		return
	}
	select {
	case l.SendTo <- relayedRequest:
		fmt.Fprintln(w, relayedRequest.Id)
	default:
		http.Error(w, "Estoy \"atorao\"", 500)
		l.PetitionStore.Delete(relayedRequest.Id)
		return
	}
}

func (l *Listener) Stop() {
	l.Stopping = true //Risky??
}
