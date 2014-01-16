package minigrush

import (
	"net/http"
	"path/filepath"

	"github.com/crbrox/store"
)

type Replyer struct {
	ReplyStore store.Interface
}

func (r *Replyer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	base := filepath.Base(req.URL.Path)
	data, e := r.ReplyStore.Get(base)
	if e != nil {
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
