// petition_test.go
package minigrush

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

var body = []byte{1, 2, 3, 0, 255, 254}

const targetHost = "veryfarawayhost:1234"
const targetScheme = "https"
const rushUrl = "http://rushhost:9876/path/subpath?q=a&r=b&r=c&españa=olé"
const relayerHostField = "x-relayer-host"
const relayerSchemeField = "x-relayer-protocol"

func TestBasic(t *testing.T) {
	fmt.Println("Hello World!")
	bytes.NewReader(body)
	original, err := http.NewRequest("GET", rushUrl, bytes.NewReader(body))
	if err != nil {
		t.Error(err)
	}
	original.Header.Set(relayerHostField, targetHost)
	original.Header.Set(relayerSchemeField, targetScheme)
	original.Header.Set("x-another-thing", "verywell")
	original.Header.Add("x-another-thing", "fandango")
	petition, err := newPetition(original)
	if err != nil {
		t.Error(err)
	}
	retrieved, err := petition.Request()
	if err != nil {
		t.Error(err)
	}
	retrievedBody, err := ioutil.ReadAll(retrieved.Body)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(body, retrievedBody) {
		t.Error("retrieved body should be equal to sent body")
	}
	//This is not necessary probably. Just in case implementation changes
	original.Header.Del(relayerHostField)
	original.Header.Del(relayerSchemeField)

	if !reflect.DeepEqual(retrieved.Header, original.Header) {
		t.Error("retrieved header should be equal to sent header")
	}
	if retrieved.URL.Host != targetHost {
		t.Errorf("retrieved host %q should be equal to target host %q", retrieved.URL.Host, targetHost)
	}
	if retrieved.URL.Scheme != targetScheme {
		t.Errorf("retrieved scheme %q should be equal to target scheme %q", retrieved.URL.Host, targetHost)
	}

	if retrieved.URL.RequestURI() != original.URL.RequestURI() {
		t.Errorf("retrieved requestURI %q should be equal to original requestURI %q",
			retrieved.URL.RequestURI(), original.URL.RequestURI())
	}

}
