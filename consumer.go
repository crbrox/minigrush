package minigrush

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/crbrox/store"
)

type Consumer struct {
	//Channel for getting petitions
	GetFrom <-chan *Petition
	//Store of petitions, for removing when done
	PetitionStore store.Interface
	//Store of replies, for saving responses
	ReplyStore store.Interface
	//http.Client for making requests to target host
	Client http.Client
	//number of goroutines consuming petitions
	n int
	//channel for asking goroutines to finish
	endChan chan bool
	//WaitGroup for goroutines after been notified the should end
	wg sync.WaitGroup
}

//Start starts n goroutines for taking Petitions from the GetFrom channel.
//It returns a channel for notifying when the consumer has ended (hopefully after a Stop() method invocation)
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

//Loop of taking a petition and making the request it represents
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

//process recreates the request that should be send to the target host
// it stores the response in the store of replies
func (c *Consumer) process(petition *Petition) {
	var (
		req   *http.Request
		resp  *http.Response
		reply *Reply
		start = time.Now()
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

//Stop asks consumer to stop taking petitions. When the stop is complete,
//the fact will be notified through the channel returned by the Start() method
func (c *Consumer) Stop() {
	for i := 0; i < c.n; i++ {
		c.endChan <- true
	}
}
