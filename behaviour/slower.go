package behaviour

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dhontecillas/reqstatsrv/config"
)

func SlowerBehaviour(next http.Handler, cfg *config.Behaviour) http.Handler {
	return NewSlower(next, SlowerConfigFromMap(cfg.Config))
}

type SlowerConfig struct {
	MaxBytesPerSecond int `json:"max_bytes_per_second"`
	FlushBytes        int `json:"flush_bytes"`
}

func SlowerConfigFromMap(m map[string]interface{}) *SlowerConfig {
	var c SlowerConfig
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

// Slower creates a delay in writing the body.
// In order to simplify the implementation, the delay
// is only applided to the body part.
type Slower struct {
	next              http.Handler
	maxBytesPerSecond int
	flushBytes        int
}

func NewSlower(next http.Handler, cnf *SlowerConfig) http.Handler {
	return &Slower{
		next:              next,
		maxBytesPerSecond: cnf.MaxBytesPerSecond,
		flushBytes:        cnf.FlushBytes,
	}
}

func (s *Slower) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	i := newSlowInstance(rw, s.maxBytesPerSecond, s.flushBytes)
	s.next.ServeHTTP(i, r)
	i.writeBody()
}

type slowInstance struct {
	rw                http.ResponseWriter
	flusher           http.Flusher
	started           time.Time
	maxBytesPerSecond int
	flushBytes        int

	delayBuffer bytes.Buffer
}

func newSlowInstance(rw http.ResponseWriter, maxBS int, flushBytes int) *slowInstance {
	return &slowInstance{
		rw:                rw,
		flusher:           rwFlusher(rw),
		started:           time.Now(),
		maxBytesPerSecond: maxBS,
		flushBytes:        flushBytes,
	}
}

func rwFlusher(rw http.ResponseWriter) http.Flusher {
	if f, ok := rw.(http.Flusher); ok {
		return f
	}
	return &nopFlusher{}
}

func (s *slowInstance) Header() http.Header {
	return s.rw.Header()
}

func (s *slowInstance) Write(b []byte) (int, error) {
	return s.delayBuffer.Write(b)
}

func (s *slowInstance) WriteHeader(statusCode int) {
	s.rw.WriteHeader(statusCode)
}

func (s *slowInstance) writeBody() {
	totalBytes := s.delayBuffer.Len()
	var written int
	sleepTime := time.Second / time.Duration(s.maxBytesPerSecond)

	var wb [1]byte
	var err error
	for s.delayBuffer.Len() > 0 {
		shouldHaveWriten := int(time.Since(s.started) / sleepTime)
		pending := shouldHaveWriten - written
		for i := 0; i < pending; i++ {
			wb[0], err = s.delayBuffer.ReadByte()
			if err != nil {
				// TODO: this should not happen
				return
			}
			s.rw.Write(wb[:1])
		}
		written += pending
		s.flusher.Flush()
		if written < totalBytes {
			time.Sleep(sleepTime)
		}
	}
}

type nopFlusher struct {
}

func (f *nopFlusher) Flush() {
}
