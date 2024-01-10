package content

import (
	"net/http"

	"github.com/dhontecillas/reqstatsrv/behaviour"
	"github.com/dhontecillas/reqstatsrv/config"
)

func EmptyContentHandler(cfg *config.Content, nestedBuilder NestedContentBuilderFn) http.Handler {
	return &EmptyPayload{}
}

type EmptyPayload struct {
}

func (d *EmptyPayload) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h := rw.Header()
	h.Add("Content-Length", "0")
	rw.WriteHeader(behaviour.ResponseStatusOr(req, 200))
	return
}
