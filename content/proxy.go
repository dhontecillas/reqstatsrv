package content

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/dhontecillas/reqstatsrv/config"
)

func ProxyContentHandler(_ *config.Endpoint, cfg *config.Content) http.Handler {
	return NewProxyContent(ProxyContentConfigFromMap(cfg.Config))
}

type ProxyContentConfig struct {
	ProxyURL string `json:"proxy_url"`
}

func ProxyContentConfigFromMap(m map[string]interface{}) *ProxyContentConfig {
	var c ProxyContentConfig
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

type ProxyContent struct {
	proxy   *httputil.ReverseProxy
	errResp []byte
}

func NewProxyContent(cfg *ProxyContentConfig) http.Handler {
	p := &ProxyContent{}
	if cfg == nil || len(cfg.ProxyURL) == 0 {
		p.errResp = []byte(`{ "err": "empty proxy url"}`)
		return p
	}

	proxyURL, err := url.Parse(cfg.ProxyURL)
	if err != nil {
		p.errResp = []byte(fmt.Sprintf(`{ "err": "%s" }`, err.Error()))
		return p
	}

	p.proxy = httputil.NewSingleHostReverseProxy(proxyURL)
	return p
}

func (c *ProxyContent) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if c.proxy == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		errResp := c.errResp
		if errResp == nil {
			errResp = []byte(`{"err": "non initialized proxy"}`)
		}
		rw.Write(errResp)
		return
	}
	c.proxy.ServeHTTP(rw, req)
}
