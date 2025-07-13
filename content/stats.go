package content

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dhontecillas/reqstatsrv/behaviour"
	"github.com/dhontecillas/reqstatsrv/config"
	"github.com/dhontecillas/reqstatsrv/stats"
)

func StatsHandler(_ *config.Endpoint, cfg *config.Content) http.Handler {
	return &CounterPayloadHandler{}
}

type CounterPayloadHandler struct{}

func (c *CounterPayloadHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	info := stats.GetStatsInfo()
	b, err := json.Marshal(info)
	if err != nil {
		b = []byte(fmt.Sprintf(`{"err": "%s"}`, err.Error()))
	}

	h := rw.Header()
	h.Set("Content-Type", "application/json")
	h.Set("Content-Length", fmt.Sprintf("%d", len(b)))
	rw.WriteHeader(behaviour.ResponseStatusOr(req, 200))
	rw.Write(b)
}
