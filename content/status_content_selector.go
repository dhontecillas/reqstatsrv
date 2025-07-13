package content

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dhontecillas/reqstatsrv/behaviour"
	"github.com/dhontecillas/reqstatsrv/config"
)

func StatusContentSelectorHandler(eCfg *config.Endpoint, cfg *config.Content) http.Handler {
	return NewStatusContentSelector(eCfg, StatusSelectorConfigFromMap(cfg.Config))
}

type StatusRangeContent struct {
	From    int            `json:"from"`
	To      int            `json:"to"`
	Content config.Content `json:"content"`
}

type StatusContentSelectorConfig struct {
	DefaultContent config.Content       `json:"default_content"`
	StatusContents []StatusRangeContent `json:"status_contents"`
}

func StatusSelectorConfigFromMap(m map[string]interface{}) *StatusContentSelectorConfig {
	var c StatusContentSelectorConfig
	// TODO: clean up and check for overlapping ranges
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

type contentByStatusRange struct {
	from int
	to   int
	next http.Handler
}

type StatusContentSelector struct {
	defaultContent http.Handler
	ranges         []contentByStatusRange
}

func NewStatusContentSelector(eCfg *config.Endpoint, cfg *StatusContentSelectorConfig) *StatusContentSelector {
	s := &StatusContentSelector{
		defaultContent: Build(eCfg, &cfg.DefaultContent),
		ranges:         make([]contentByStatusRange, 0, len(cfg.StatusContents)),
	}
	for _, rg := range cfg.StatusContents {
		s.ranges = append(s.ranges, contentByStatusRange{
			from: rg.From,
			to:   rg.To,
			next: Build(eCfg, &rg.Content),
		})
	}
	return s
}

func (d *StatusContentSelector) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	status := behaviour.ResponseStatusOr(req, 200)
	for _, sc := range d.ranges {
		if status >= sc.from && status < sc.to {
			sc.next.ServeHTTP(rw, req)
			return
		}
	}
	d.defaultContent.ServeHTTP(rw, req)
}
