package stats

import (
	"sync"
	"time"
)

var (
	globalCounters *Counters
	globalInit     sync.Once

	endpointStats    []*EndpointStats
	endpointStatsMux sync.Mutex
)

func GetGlobalCounters() *Counters {
	globalInit.Do(func() {
		globalCounters = &Counters{
			startsAt: time.Now().Unix(),
		}
	})
	return globalCounters
}

func NewEndpointStats(endpoint string) *EndpointStats {
	es := &EndpointStats{
		endpoint: endpoint,
		counters: NewCounters(),
	}
	endpointStatsMux.Lock()
	endpointStats = append(endpointStats, es)
	endpointStatsMux.Unlock()
	return es
}

type EndpointStatsInfo struct {
	Endpoint string `json:"endpoint"`
	Count    int64  `json:"count"`
}

type StatsInfo struct {
	Total     int64               `json:"total"`
	Endpoints []EndpointStatsInfo `json:"endpoints"`
}

func GetStatsInfo() *StatsInfo {
	endpointStatsMux.Lock()
	epis := make([]EndpointStatsInfo, 0, len(endpointStats))

	for _, es := range endpointStats {
		epis = append(epis, EndpointStatsInfo{
			Endpoint: es.endpoint,
			Count:    es.counters.InstantCount(),
		})
	}
	endpointStatsMux.Unlock()

	return &StatsInfo{
		Total:     globalCounters.InstantCount(),
		Endpoints: epis,
	}
}
