package behaviour

import (
	"fmt"
	"net/http"

	"github.com/dhontecillas/reqstatsrv/config"
)

type BehaviourHandlerBuilderFn func(next http.Handler, cfg *config.Behaviour) http.Handler

var (
	behaviourBuilders = map[string]BehaviourHandlerBuilderFn{
		"connection_closer":  ConnectionCloserBehaviour,
		"delayer":            DelayerBehaviour,
		"slower":             SlowerBehaviour,
		"status_distributor": StatusDistributorBehaviour,
	}
)

func Build(contentHandler http.Handler, behaviours []config.Behaviour) http.Handler {
	next := contentHandler
	for _, cfg := range behaviours {
		if b, ok := behaviourBuilders[cfg.Name]; ok {
			h := b(next, &cfg)
			if h != nil {
				next = h
			} else {
				fmt.Printf("cannot apply behaviour %q\n", cfg.Name)
			}
		} else {
			fmt.Printf("behavior builder %q not found: falling back to empty content\n", cfg.Name)
		}
	}
	return next
}
