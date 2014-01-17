package minigrush

import (
	"io/ioutil"
	"net/http"
	"time"
)

//Reply represents the response from the target host
type Reply struct {
	//Reply id. Currently the same as the petition id
	Id string
	//Possible error in making the request. Could be ""
	Error      string
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	Header     http.Header
	Trailer    http.Header
	Body       []byte
	//Petition that
	Petition *Petition
	//Beginning of the request
	Created time.Time
	//Time when response was received
	Done time.Time
}

//newReply returns the Reply for the Petition made, the http.Response got and the possible error
func newReply(resp *http.Response, p *Petition, e error) *Reply {
	var reply = &Reply{Id: p.Id, Petition: p}
	if e != nil {
		reply.Error = e.Error()
		return reply
	}
	reply.StatusCode = resp.StatusCode
	reply.Proto = resp.Proto
	reply.Header = resp.Header
	reply.Trailer = resp.Trailer
	body, err := ioutil.ReadAll(resp.Body)
	reply.Done = time.Now()
	if err != nil {
		reply.Error = e.Error()
	} else {
		reply.Body = body
	}
	return reply
}
