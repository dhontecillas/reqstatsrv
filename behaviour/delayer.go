package behaviour

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dhontecillas/reqstatsrv/config"
	"github.com/dhontecillas/reqstatsrv/stats"
)

func DelayerBehaviour(next http.Handler, cfg *config.Behaviour) http.Handler {
	return NewDelayer(next, DelayerConfigFromMap(cfg.Config))
}

type DelayerConfig struct {
	DelayMillisDistribution config.IntDistribution `json:"delay_millis_distribution"`
	Seed                    int64                  `json:"seed"`
}

func DelayerConfigFromMap(m map[string]interface{}) *DelayerConfig {
	var c DelayerConfig
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

type Delayer struct {
	next          http.Handler
	distrInstance stats.IntDistributionInstance
}

func NewDelayer(next http.Handler, cfg *DelayerConfig) http.Handler {
	distrcfg := stats.NewIntDistribution(cfg.DelayMillisDistribution)
	distrInstance := distrcfg.Instance(cfg.Seed)

	// TODO: add the `OnBytesWritten` implementation
	return &Delayer{
		next:          next,
		distrInstance: distrInstance,
	}
}

func (d *Delayer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	delayMillis := d.distrInstance.GetInterpolated()
	fmt.Printf("delaying %d millis\n", delayMillis)
	time.Sleep(time.Millisecond * time.Duration(delayMillis))
	d.next.ServeHTTP(rw, r)
}
