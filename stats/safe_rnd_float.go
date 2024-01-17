package stats

import (
	"math/rand"
	"sync"
)

type SafeRnd interface {
	F64() float64
}

func NewSafeRnd(seed int64) SafeRnd {
	source := rand.NewSource(seed)
	rnd := rand.New(source)

	return &safeRnd{
		rnd: rnd,
	}
}

type safeRnd struct {
	rnd   *rand.Rand
	mutex sync.Mutex
}

func (r *safeRnd) F64() float64 {
	r.mutex.Lock()
	v := r.rnd.Float64()
	r.mutex.Unlock()
	return v
}
