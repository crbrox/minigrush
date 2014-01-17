package minigrush

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/crbrox/store"
)

//Listener is responsible for receiving requests and storing them in the store associated with it.
//It then passes a reference to the object Petition which wraps the original HTTP request through the channel Sendto,
//where the Consumer should collected it for further processing
type Listener struct {
	//Channel for sending petitions
	SendTo chan<- *Petition
	//Store for saving petitions in case of crash
	PetitionStore store.Interface
	//Flag signaling listener should finish
	stopping bool
}

//ServeHTTP implements HTTP handler interface
func (l *Listener) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if l.stopping {
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

//Stop asks listener to stop receiving petitions
func (l *Listener) Stop() {
	l.stopping = true //Risky??
}
