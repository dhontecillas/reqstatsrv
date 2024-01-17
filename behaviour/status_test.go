package behaviour

import (
	"strings"
	"testing"
)

func TestBehaviourStatus_Happy(t *testing.T) {
	s := StatusDistributorConfig{
		CodeDistribution: map[int]float64{
			200: 0.9,
			500: 0.1,
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
		CodeDistribution: map[int]float64{},
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
		CodeDistribution: map[int]float64{
			200: 120,
			500: 240,
		},
	}

	err := s.Clean()
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}

	v200 := s.CodeDistribution[200]
	if v200 <= 0.33 || v200 >= 0.34 {
		t.Errorf("expected 200 value 0.3333.. got %f", v200)
		return
	}
	v500 := s.CodeDistribution[500]
	if v500 <= 0.66 || v500 >= 0.67 {
		t.Errorf("expected 200 value 0.6666.. got %f", v500)
		return
	}
}
