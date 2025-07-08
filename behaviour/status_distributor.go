package behaviour

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dhontecillas/reqstatsrv/config"
	"github.com/dhontecillas/reqstatsrv/stats"
)

func StatusDistributorBehaviour(next http.Handler, cfg *config.Behaviour) http.Handler {
	return NewStatusDistributor(next, StatusDistributorConfigFromMap(cfg.Config))
}

type StatusDistributorConfig struct {
	CodeDistribution config.IntDistribution `json:"code_distribution"`
	Seed             int64                  `json:"seed"`
}

func StatusDistributorConfigFromMap(m map[string]interface{}) *StatusDistributorConfig {
	fmt.Printf("status distributor: %#v \n", m)
	var c StatusDistributorConfig
	b, err := json.Marshal(m)
	if err != nil {
		fmt.Printf("error %s converting map %#v to json\n", err.Error(), m)
		return &c
	}
	err = json.Unmarshal(b, &c)
	if err != nil {
		fmt.Printf("error %s creating config from %s\n", err.Error(), string(b))
	}
	fmt.Printf("status distributor parsed: %#v \n", c)
	c.Clean()
	return &c
}

// Clean checks that all the provided numbers are valid status
// codes, and updates the `CodeDistribution` field to only hold
// the valid ones. It returns an error if the CodeDistribution
// keys had changed, explaining the reason.
// Normalizing the frequency values is not considered a "modification",
// so no error will be returned.
// If an empty CodeDistribution is provided it will default to a 200 Ok
// with 100% probability.
func (s *StatusDistributorConfig) Clean() error {
	var errB strings.Builder

	dups := map[int]bool{}
	clean := make(config.IntDistribution, 0, len(s.CodeDistribution))
	for _, kv := range s.CodeDistribution {
		if kv.Val <= 0.0 {
			errB.WriteString(fmt.Sprintf("invalid value %f for %d;", kv.Val, kv.Key))
			continue
		}

		if http.StatusText(kv.Key) == "" {
			// not valid status code (is a "custom" one)
			errB.WriteString(fmt.Sprintf("invalid code %d (not known);", kv.Key))
			continue
		}
		if dups[kv.Key] {
			errB.WriteString(fmt.Sprintf("duplicate for %d wit value %f;", kv.Key, kv.Val))
		} else {
			clean = append(clean, kv)
			dups[kv.Key] = true
		}
	}

	if len(clean) == 0 {
		clean = []config.IntFloat{
			{Key: 200, Val: 1.0},
		}
		errB.WriteString("empty code distribution: falling back to 200 Ok always;")
	}

	s.CodeDistribution = clean

	if errB.Len() == 0 {
		return nil
	}

	return errors.New(errB.String())
}

type responseStatusType string

var responseStatus = responseStatusType("response_status")

func WithResponseStatus(req *http.Request, status int) *http.Request {
	ctx := context.WithValue(req.Context(), responseStatus, status)
	return req.WithContext(ctx)
}

func ResponseStatus(req *http.Request) int {
	v := req.Context().Value(responseStatus)
	if v == nil {
		return 0
	}
	if s, ok := v.(int); ok {
		return s
	}
	return 0
}

func ResponseStatusOr(req *http.Request, fallback int) int {
	s := ResponseStatus(req)
	if s == 0 {
		return fallback
	}
	return s
}

type StatusDistributionSelector struct {
	distrInstance stats.IntDistributionInstance
	next          http.Handler
}

func NewStatusDistributor(next http.Handler, cnf *StatusDistributorConfig) *StatusDistributionSelector {
	cnf.Clean()
	distrCnf := stats.NewIntDistribution(cnf.CodeDistribution)
	distrInstance := distrCnf.Instance(cnf.Seed)
	return &StatusDistributionSelector{
		distrInstance: distrInstance,
		next:          next,
	}
}

func (s *StatusDistributionSelector) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// we just add the status code to return to the context, so payload generation
	// function can read if it is set, and select the kind of payload depending on the status
	// code.
	status_code := s.distrInstance.Get()
	rr := WithResponseStatus(r, status_code)
	s.next.ServeHTTP(rw, rr)
}
