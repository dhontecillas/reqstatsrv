package stats

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	// A full day of timings
	MaxSeconds int64 = 60 * 60 * 24
)

type Counters struct {
	startsAt int64
	seconds  [60 * 60 * 24]int64
}

func NewCounters() *Counters {
	n := time.Now().Unix()
	c := &Counters{
		startsAt: n,
	}
	return c
}

func (c *Counters) Inc() error {
	bucketIdx := time.Now().Unix() - c.startsAt
	if bucketIdx >= MaxSeconds {
		// after a full day running the stats server.. we just stop counting,
		return fmt.Errorf("max tracking time reached")
	}
	atomic.AddInt64(&(c.seconds[bucketIdx]), 1)
	return nil
}

func (c *Counters) InstantCount() int64 {
	bucketIdx := time.Now().Unix() - c.startsAt
	if bucketIdx >= MaxSeconds {
		bucketIdx = MaxSeconds
	}
	// this is to return the known rate from a "completed" second.
	// get the information from the previous second
	if bucketIdx != 0 {
		bucketIdx -= 1
	}
	return atomic.LoadInt64(&(c.seconds[bucketIdx]))
}

type EndpointStats struct {
	endpoint string
	counters *Counters
}

type CounterHandler struct {
	next     http.Handler
	global   *Counters
	endpoint *EndpointStats
}

func NewStatsMiddleware(next http.Handler, endpoint string,
	endpointStats *EndpointStats, globalCounters *Counters) *CounterHandler {
	if endpointStats == nil {
		endpointStats = NewEndpointStats(endpoint)
	}
	if globalCounters == nil {
		globalCounters = GetGlobalCounters()
	}
	return &CounterHandler{
		next:     next,
		global:   globalCounters,
		endpoint: endpointStats,
	}
}

func (h *CounterHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.global.Inc()
	h.endpoint.counters.Inc()
	h.next.ServeHTTP(rw, req)
}
