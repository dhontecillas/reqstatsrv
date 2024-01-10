package behaviour

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dhontecillas/reqstatsrv/config"
	"github.com/dhontecillas/reqstatsrv/stats"
)

func ConnectionCloserBehaviour(next http.Handler, cfg *config.Behaviour) http.Handler {
	return NewConnectionCloser(next, ConnectionCloserConfigFromMap(cfg.Config))
}

type ConnectionCloserConfig struct {
	Freq float64 `json:"freq"`
	Seed int64   `json:"seed"`
}

func ConnectionCloserConfigFromMap(m map[string]interface{}) *ConnectionCloserConfig {
	var c ConnectionCloserConfig
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

type ConnectionCloser struct {
	next http.Handler
	rnd  stats.SafeRnd
	freq float64
}

func NewConnectionCloser(next http.Handler, cfg *ConnectionCloserConfig) http.Handler {
	if cfg.Freq < 0.0 {
		return next
	}

	freq := cfg.Freq
	if cfg.Freq > 1.0 {
		freq = 1.0
	}

	// TODO: add the `OnBytesWritten` implementation
	return &ConnectionCloser{
		next: next,
		rnd:  stats.NewSafeRnd(cfg.Seed),
		freq: freq,
	}
}

func (s *ConnectionCloser) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	closeConn := s.rnd.F64() < s.freq
	if !closeConn {
		s.next.ServeHTTP(rw, r)
		return
	}

	if h, ok := rw.(http.Hijacker); ok {
		conn, _, err := h.Hijack()
		if conn != nil {
			conn.Close()
		} else {
			if err != nil {
				fmt.Printf("cannot hijack connection: %s\n", err.Error())
			}
			// we cannot drop the connection
			return
		}
	}

}
