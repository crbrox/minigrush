package minigrush

import (
	"io/ioutil"
	"net/http"
	"time"
)

type Reply struct {
	Id         string
	Error      string
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	Header     http.Header
	Trailer    http.Header
	Body       []byte
	Petition   *Petition
	Created    time.Time
	Done       time.Time
}

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
