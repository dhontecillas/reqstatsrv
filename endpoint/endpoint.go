package endpoint

import (
	"net/http"

	"github.com/dhontecillas/reqstatsrv/behaviour"
	"github.com/dhontecillas/reqstatsrv/config"
	"github.com/dhontecillas/reqstatsrv/content"
	"github.com/dhontecillas/reqstatsrv/stats"
)

func Bind(mux *http.ServeMux, cfg *config.Endpoint) {
	ch := content.Build(cfg, &cfg.Content)
	h := behaviour.Build(ch, cfg.Behaviour)

	s := stats.NewStatsMiddleware(h, cfg.PathPattern, nil, nil)
	if cfg.Method == "" {
		mux.Handle(cfg.PathPattern, s)
	} else {
		mux.Handle(cfg.Method+" "+cfg.PathPattern, s)
	}
}
