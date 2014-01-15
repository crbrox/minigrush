package grass

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/crbrox/store"
)

type Consumer struct {
	GetFrom       <-chan *Petition
	PetitionStore store.Interface
	ReplyStore    store.Interface
	Client        http.Client
	n             int
	endChan       chan bool
	wg            sync.WaitGroup
}

func (c *Consumer) Start(n int) <-chan bool {
	c.n = n
	finalDone := make(chan bool)
	c.endChan = make(chan bool, c.n)
	c.wg.Add(c.n)
	for i := 0; i < c.n; i++ {
		go c.relay()
	}
	go func() {
		c.wg.Wait()
		finalDone <- true
	}()
	return finalDone
}

func (c *Consumer) relay() {
	defer c.wg.Done()
SERVE:
	for {
		select {
		case <-c.endChan:
			break SERVE
		default:
			select {
			case req := <-c.GetFrom:
				c.process(req)
			case <-c.endChan:
				break SERVE
			}
		}
	}
}

func (c *Consumer) process(petition *Petition) {
	var (
		req   *http.Request
		resp  *http.Response
		reply *Reply
		start time.Time = time.Now()
	)
	req, err := petition.Request()
	if err != nil {
		log.Println(petition.Id, err)
	} else {
		resp, err = c.Client.Do(req)
		if err != nil {
			log.Println(petition.Id, err)

		} else {
			defer resp.Body.Close()
		}
	}
	reply = newReply(resp, petition, err)
	reply.Created = start
	text, err := json.MarshalIndent(reply, "", " ")
	if err != nil {
		log.Println(petition.Id, err)
	}
	err = c.ReplyStore.Put(reply.Id, text)
	if err != nil {
		log.Println(petition.Id, err)
	}

	err = c.PetitionStore.Delete(petition.Id)
	if err != nil {
		log.Println(petition.Id, err)
	}

}

func (c *Consumer) Stop() {
	for i := 0; i < c.n; i++ {
		c.endChan <- true
	}
}
