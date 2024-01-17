package stats

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"

	"github.com/dhontecillas/reqstatsrv/config"
)

func NormalizeIntDistribution(in map[int]float64) map[int]float64 {
	var sum float64
	for _, v := range in {
		if v < 0.0 {
			continue
		}
		sum += v
	}
	out := make(map[int]float64, len(in))
	for k, v := range in {
		if v < 0.0 {
			continue
		}
		out[k] = v / sum
	}
	return out
}

type IntDistribution struct {
	vals  []int
	distr []float64
}

type IntDistributionInstance interface {
	Get() int
	GetInterpolated() int
}

func NewIntDistributionFromMap(m map[int]float64) *IntDistribution {
	if len(m) == 0 {
		return &IntDistribution{}
	}
	vals := make([]int, 0, len(m))
	distr := make([]float64, 0, len(m))
	var sum float64

	nm := NormalizeIntDistribution(m)

	// in order to have interpolated values, we want the values
	// sorted, so we can interpolate "values ramp"
	sortedKeys := sort.IntSlice(make([]int, 0, len(nm)))
	for k, _ := range nm {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Sort(sortedKeys)

	for _, k := range sortedKeys {
		v := nm[k]
		vals = append(vals, k)
		sum += v
		distr = append(distr, sum)
	}

	if sum != 1.0 {
		// avoid rounding errors that migh arise from
		// the "normalization" of values.
		distr[len(distr)-1] = 1.0
	}

	return &IntDistribution{
		vals:  vals,
		distr: distr,
	}
}

func NewIntDistribution(d config.IntDistribution) *IntDistribution {
	var err error
	m := make(map[int]float64, len(d))
	for _, kv := range d {
		if _, ok := m[kv.Key]; ok {
			err = fmt.Errorf("has duplicate keys: last value will prevail")
		}
		m[kv.Key] = kv.Val
	}
	if err != nil {
		fmt.Printf("NewIntDistribution err: %s", err.Error())
	}
	return NewIntDistributionFromMap(m)
}

func (c *IntDistribution) Val(in float64) int {
	if len(c.distr) == 0 {
		return 0
	}

	if in < 0.0 {
		// in = 0.0
		return c.vals[0]
	}
	if in > 1.0 {
		// in = 1.0
		return c.vals[len(c.vals)-1]
	}

	for idx, upper := range c.distr {
		if upper > in {
			fmt.Printf("vals: %#v, distr: %#v -> %d\n", c.vals, c.distr, c.vals[idx])
			return c.vals[idx]
		}
	}
	return c.vals[len(c.vals)-1]
}

func (c *IntDistribution) InterpolatedVal(in float64) int {
	if len(c.distr) == 0 {
		return 0
	}
	if in < 0.0 {
		// in = 0.0
		return c.vals[0]
	}
	if in > 1.0 {
		// in = 1.0
		return c.vals[len(c.vals)-1]
	}

	prevVal := int(0)
	prevUpper := float64(0.0)
	for idx, upper := range c.distr {
		v := c.vals[idx]
		if upper > in {
			k := (in - prevUpper) / (upper - prevUpper)
			rval := prevVal + int(math.Round(float64(v-prevVal)*k))
			fmt.Printf("vals: %#v, distr: %#v -> %d\n", c.vals, c.distr, c.vals[idx])
			return rval
		}
		prevVal = v
		prevUpper = upper
	}
	return c.vals[len(c.vals)-1]
}

func (d *IntDistribution) Instance(seed int64) IntDistributionInstance {
	return newIntDistrInstance(seed, d)
}

type intDistrInstance struct {
	rnd   *rand.Rand
	mutex sync.Mutex

	distr *IntDistribution
}

func newIntDistrInstance(seed int64, distr *IntDistribution) *intDistrInstance {
	source := rand.NewSource(seed)
	rnd := rand.New(source)
	return &intDistrInstance{
		rnd:   rnd,
		distr: distr,
	}
}

func (d *intDistrInstance) Get() int {
	d.mutex.Lock()
	v := d.rnd.Float64()
	d.mutex.Unlock()
	fmt.Printf("val: %f\n", v)
	return d.distr.Val(v)
}

func (d *intDistrInstance) GetInterpolated() int {
	d.mutex.Lock()
	v := d.rnd.Float64()
	d.mutex.Unlock()
	return d.distr.InterpolatedVal(v)
}
