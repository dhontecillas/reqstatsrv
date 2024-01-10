package content

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/dhontecillas/reqstatsrv/behaviour"
	"github.com/dhontecillas/reqstatsrv/config"
)

func FileContentHandler(cfg *config.Content, nestedBuilder NestedContentBuilderFn) http.Handler {
	return NewFileContent(FileContentConfigFromMap(cfg.Config))
}

type FileContentConfig struct {
	Path string `json:"path"`
}

func FileContentConfigFromMap(m map[string]interface{}) *FileContentConfig {
	var c FileContentConfig
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

type FileContent struct {
	content       []byte
	contentType   string
	contentLength string
}

func NewFileContent(cfg *FileContentConfig) http.Handler {
	b, rErr := os.ReadFile(cfg.Path)
	if rErr != nil {
		fmt.Printf("cannot read content from file %s: %s\n", cfg.Path, rErr.Error())
		return &FileContent{}
	}

	return &FileContent{
		content:       b,
		contentType:   getContentTypeFromExtension(cfg.Path),
		contentLength: fmt.Sprintf("%d", len(b)),
	}
}

func (c *FileContent) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if len(c.contentLength) == 0 {
		rw.WriteHeader(behaviour.ResponseStatusOr(req, 200))
		return
	}

	h := rw.Header()
	if c.contentType != "" {
		h.Set("Content-Type", c.contentType)
	}
	h.Set("Content-Length", c.contentLength)
	rw.WriteHeader(behaviour.ResponseStatusOr(req, 200))
	rw.Write(c.content)
}
