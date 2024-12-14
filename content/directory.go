package content

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/dhontecillas/reqstatsrv/behaviour"
	"github.com/dhontecillas/reqstatsrv/config"
)

func DirectoryContentHandler(cfg *config.Content, nestedBuilder NestedContentBuilderFn) http.Handler {
	return NewDirectoryContent(DirectoryContentConfigFromMap(cfg.Config))
}

// DirectoryContentConfig allows to define how to find the fake files
// inside a directory.
//
// AttemptExtensions allows to define a set of extension to append, when
// the path in the url does not match a file.
//
// DunderQueryStrings option, sorts the query params alphabetically, and
// joins them after the file using "dunder" (double underscore) separators
// between key, value pairs. Key Value pairs are split by using a single
// underscore. So a path like: /foo/bar?a=foo&b=bar will become:
// /foo/bar__a_foo__b_bar
type DirectoryContentConfig struct {
	Dir                string   `json:"dir"`
	EndpointPath       string   `json:"endpoint_path"`
	AttemptExtensions  []string `json:"attempt_extensions"`
	DunderQueryStrings bool     `json:"dunder_querystrings"`
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
	dir                http.Dir
	parentPath         string
	attemptExtensions  []string
	dunderQueryStrings bool
}

func NewDirectoryContent(cfg *DirectoryContentConfig) http.Handler {
	// the problem with FileServer is that the status code from
	// the context is not respected
	return &DirectoryContent{
		dir:                http.Dir(cfg.Dir),
		parentPath:         cfg.EndpointPath,
		attemptExtensions:  cfg.AttemptExtensions,
		dunderQueryStrings: cfg.DunderQueryStrings,
	}
}

func (c *DirectoryContent) findFile(req *http.Request) (http.File, error) {
	// we remove the final `/` if present, and remove '..' / '.' path elements
	p := path.Clean(req.URL.Path)

	// TODO: apply the Dunder config option to be able to serve different
	// files based on the query strings

	f, err := c.dir.Open(p)
	if err == nil {
		s, serr := f.Stat()
		if serr == nil && !s.IsDir() {
			return f, err
		}
	}

	// attempt with any of the allowed extensions
	for _, ext := range c.attemptExtensions {
		if f, err = c.dir.Open(p + "." + ext); err == nil {
			return f, err
		}
	}

	return nil, fmt.Errorf("not found")
}

func (c *DirectoryContent) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	f, err := c.findFile(req)
	if err != nil {
		rw.WriteHeader(behaviour.ResponseStatusOr(req, 404))
		return
	}

	b, rerr := io.ReadAll(f)
	if rerr != nil {
		rw.WriteHeader(behaviour.ResponseStatusOr(req, 503))
		return
	}

	h := rw.Header()
	h.Set("Content-Length", fmt.Sprintf("%d", len(b)))
	// try to infer the content type of the file
	var ct string
	if s, err := f.Stat(); err != nil {
		ct = getContentTypeFromExtension(s.Name())
	} else {
		ct = getContentTypeFromExtension(req.URL.Path)
	}
	if ct != "" {
		h.Set("Content-Type", ct)
	}

	rw.WriteHeader(behaviour.ResponseStatusOr(req, 200))
	rw.Write(b)
}
