package content

import (
	"net/http"

	"github.com/dhontecillas/reqstatsrv/behaviour"
	"github.com/dhontecillas/reqstatsrv/config"
)

func EmptyContentHandler(_ *config.Endpoint, cfg *config.Content) http.Handler {
	return &EmptyPayload{}
}

type EmptyPayload struct {
}

func (d *EmptyPayload) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h := rw.Header()
	h.Add("Content-Length", "0")
	// if there is no content, the default http status code should be 204
	rw.WriteHeader(behaviour.ResponseStatusOr(req, http.StatusNoContent))
}
