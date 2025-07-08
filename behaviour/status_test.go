package behaviour

import (
	"strings"
	"testing"

	"github.com/dhontecillas/reqstatsrv/config"
)

func TestBehaviourStatus_Happy(t *testing.T) {
	s := StatusDistributorConfig{
		CodeDistribution: config.IntDistribution{
			config.IntFloat{Key: 200, Val: 0.9},
			config.IntFloat{Key: 500, Val: 0.1},
		},
	}

	err := s.Clean()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
}

func TestBehaviourStatus_FallbackToDefault(t *testing.T) {
	s := StatusDistributorConfig{
		CodeDistribution: config.IntDistribution{},
	}

	err := s.Clean()
	if err == nil {
		t.Errorf("expected error: got nil")
		return
	}

	if !strings.Contains(err.Error(), "falling back") {
		t.Errorf("expecting 'falling back' error message")
		return
	}
}

func TestBehaviourStatus_Normalize(t *testing.T) {
	s := StatusDistributorConfig{
		CodeDistribution: config.IntDistribution{
			config.IntFloat{Key: 200, Val: 120.0},
			config.IntFloat{Key: 500, Val: 240.0},
		},
	}

	err := s.Clean()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	v200 := s.CodeDistribution[0].Val
	if v200 <= 0.33 || v200 >= 0.34 {
		t.Errorf("expected 200 value 0.3333.. got %f", v200)
		return
	}
	v500 := s.CodeDistribution[1].Val
	if v500 <= 0.66 || v500 >= 0.67 {
		t.Errorf("expected 200 value 0.6666.. got %f", v500)
		return
	}
}
