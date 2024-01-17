package content

import (
	"net/http"

	"github.com/dhontecillas/reqstatsrv/behaviour"
	"github.com/dhontecillas/reqstatsrv/config"
)

func DummyHandler(cfg *config.Content, nestedBuilder NestedContentBuilderFn) http.Handler {
	return &DummyPayload{}
}

type DummyPayload struct {
}

func (d *DummyPayload) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("My-Test", "foo")
	rw.WriteHeader(behaviour.ResponseStatusOr(req, 200))
	rw.Write([]byte("\nthis\nis\nsomething"))
}
