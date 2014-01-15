package grass

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"
)

const (
	RelayerHost     = "X-Relayer-Host"
	RelayerProtocol = "X-Relayer-Protocol"
)

type Petition struct {
	Id           string
	TargetHost   string
	TargetScheme string
	Method       string // GET, POST, PUT, etc.
	URL          *url.URL
	URLString    string
	Proto        string // "HTTP/1.0"
	Header       http.Header
	Trailer      http.Header
	Body         []byte
	RemoteAddr   string
	RequestURI   string
	Host         string
	Created      time.Time
}

func newPetition(original *http.Request) (*Petition, error) {
	targetHost := original.Header.Get(RelayerHost)
	if targetHost == "" {
		return nil, fmt.Errorf("Missing mandatory header %s", RelayerHost)
	}
	original.Header.Del(RelayerHost)
	scheme := strings.ToLower(original.Header.Get(RelayerProtocol))
	switch scheme {
	case "http", "https":
	case "":
		scheme = "http"
	default:
		return nil, fmt.Errorf("Unsupported protocol %s", scheme)

	}
	original.Header.Del(RelayerProtocol)
	//save body content
	body, err := ioutil.ReadAll(original.Body)
	if err != nil {
		return nil, err
	}
	id := uuid.New()
	relayedRequest := &Petition{
		Id:           id,
		Body:         body,
		Method:       original.Method,
		URL:          original.URL,
		Proto:        original.Proto, // "HTTP/1.0"
		Header:       original.Header,
		Trailer:      original.Trailer,
		RemoteAddr:   original.RemoteAddr,
		RequestURI:   original.RequestURI,
		TargetHost:   targetHost,
		TargetScheme: scheme,
		Created:      time.Now()}
	return relayedRequest, nil
}
func (p *Petition) Request() (*http.Request, error) {
	p.URL.Host = p.TargetHost
	p.URL.Scheme = p.TargetScheme
	p.URLString = p.URL.String()
	req, err := http.NewRequest(
		p.Method,
		p.URLString,
		ioutil.NopCloser(bytes.NewReader(p.Body))) //Restore body as ReadCloser
	if err != nil {
		return nil, err
	}
	req.Header = p.Header
	req.Trailer = p.Trailer

	return req, nil
}
