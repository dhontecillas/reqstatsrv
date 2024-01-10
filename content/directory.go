package content

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dhontecillas/reqstatsrv/behaviour"
	"github.com/dhontecillas/reqstatsrv/config"
)

func DirectoryContentHandler(cfg *config.Content, nestedBuilder NestedContentBuilderFn) http.Handler {
	return NewDirectoryContent(DirectoryContentConfigFromMap(cfg.Config))
}

type DirectoryContentConfig struct {
	Dir          string `json:"dir"`
	EndpointPath string `json:"endpoint_path"`
}

func DirectoryContentConfigFromMap(m map[string]interface{}) *DirectoryContentConfig {
	var c DirectoryContentConfig
	b, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("error %s converting map %#v to json\n", err.Error(), m)
		return &c
	}
	err = json.Unmarshal(b, &c)
	if err != nil {
		fmt.Printf("error %s creating config from %s\n", err.Error(), string(b))
	}
	return &c
}

type DirectoryContent struct {
	dir        http.Dir
	parentPath string
}

func NewDirectoryContent(cfg *DirectoryContentConfig) http.Handler {
	// the problem with FileServer is that the status code from
	// the context is not respected
	return &DirectoryContent{
		dir:        http.Dir(cfg.Dir),
		parentPath: cfg.EndpointPath,
	}
}

func (c *DirectoryContent) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if f, err := c.dir.Open(req.URL.Path); err == nil {
		if b, rerr := io.ReadAll(f); rerr == nil {
			h := rw.Header()
			h.Set("Content-Length", fmt.Sprintf("%d", len(b)))
			if ct := getContentTypeFromExtension(req.URL.Path); ct != "" {
				h.Set("Content-Type", ct)
			}
			rw.WriteHeader(behaviour.ResponseStatusOr(req, 200))
			rw.Write(b)
			return
		}
	}
	rw.WriteHeader(behaviour.ResponseStatusOr(req, 200))
}
