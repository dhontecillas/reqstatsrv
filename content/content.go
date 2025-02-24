package content

import (
	"fmt"
	"net/http"

	"github.com/dhontecillas/reqstatsrv/config"
)

type NestedContentBuilderFn func(c *config.Content) http.Handler

type ContentHandlerBuilderFn func(c *config.Content, nestedBuilder NestedContentBuilderFn) http.Handler

var (
	contentBuilders = map[string]ContentHandlerBuilderFn{
		"directory":               DirectoryContentHandler,
		"file":                    FileContentHandler,
		"empty":                   EmptyContentHandler,
		"dummy":                   DummyHandler,
		"stats":                   StatsHandler,
		"status_content_selector": StatusContentSelectorHandler,
		"proxy":                   ProxyContentHandler,
	}
)

func Build(cfg *config.Content) http.Handler {
	if b, ok := contentBuilders[cfg.Source]; ok {
		h := b(cfg, Build)
		if h != nil {
			return h
		}
		fmt.Printf("cannot build '%s': falling back to empty content", cfg.Source)
	} else {
		fmt.Printf("content builder '%s' not found: falling back to empty content\n", cfg.Source)
	}
	return &EmptyPayload{}
}
