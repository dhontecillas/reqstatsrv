package content

import (
	"fmt"
	"net/http"

	"github.com/dhontecillas/reqstatsrv/config"
)

type NestedContentBuilderFn func(c *config.Content) http.Handler

type ContentHandlerBuilderFn func(endpointCfg *config.Endpoint,
	c *config.Content) http.Handler

func Build(eCfg *config.Endpoint, cfg *config.Content) http.Handler {
	contentBuilders := map[string]ContentHandlerBuilderFn{
		"directory":               DirectoryContentHandler,
		"file":                    FileContentHandler,
		"empty":                   EmptyContentHandler,
		"dummy":                   DummyHandler,
		"stats":                   StatsHandler,
		"status_content_selector": StatusContentSelectorHandler,
		"proxy":                   ProxyContentHandler,
	}
	if b, ok := contentBuilders[cfg.Source]; ok {
		h := b(eCfg, cfg)
		if h != nil {
			return h
		}
		fmt.Printf("cannot build '%s': falling back to empty content", cfg.Source)
	} else {
		fmt.Printf("content builder '%s' not found: falling back to empty content\n", cfg.Source)
	}
	return &EmptyPayload{}
}
