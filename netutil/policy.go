package netutil

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"net/http"
	"sync"
)

var (
	_ Balancer = &reverseRandom{}
	_ Balancer = &reverseRoundRobin{}

	errBackendsRequired = fmt.Errorf("reverse proxy backends required, got 0")
)

func Random(backends []*ReverseProxyBackend) (*reverseRandom, error) {
	if len(backends) == 0 {
		return nil, errBackendsRequired
	}

	weightN := 0
	weights := make([]int, len(backends))

	for i, backend := range backends {
		weightN += backend.Weight
		weights[i] = weightN
	}

	return &reverseRandom{
		backends: backends,

		weightN: weightN,
		weights: weights,
	}, nil
}

type reverseRandom struct {
	backends []*ReverseProxyBackend

	weightN int
	weights []int
}

func (this *reverseRandom) Pick(req *http.Request) http.Handler {
	if len(this.backends) == 1 {
		return this.backends[0]
	}

	rnd := rand.Intn(this.weightN)
	for idx, n := range this.weights {
		if rnd < n {
			return this.backends[idx]
		}
	}

	return this.backends[rand.Intn(len(this.backends))]
}

func (this *reverseRandom) String() string {
	return fmt.Sprintf("[RANDOM] %d backends, weighs %v", len(this.backends), this.weights)
}

func RoundRobin(backends []*ReverseProxyBackend) (*reverseRoundRobin, error) {
	if len(backends) == 0 {
		return nil, errBackendsRequired
	}

	return &reverseRoundRobin{
		backends: backends,
	}, nil
}

type reverseRoundRobin struct {
	backends []*ReverseProxyBackend

	mutex sync.Mutex
	idx   int
}

func (this *reverseRoundRobin) Pick(req *http.Request) http.Handler {
	if len(this.backends) == 1 {
		return this.backends[0]
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	h := this.backends[this.idx]

	this.idx = (this.idx + 1) % len(this.backends)
	return h
}

func (this *reverseRoundRobin) String() string {
	return fmt.Sprintf("[ROUND ROBIN] %d backends", len(this.backends))
}

func Hash(backends []*ReverseProxyBackend) (*reverseHash, error) {
	if len(backends) == 0 {
		return nil, errBackendsRequired
	}

	return &reverseHash{
		backends: backends,
	}, nil
}

type reverseHash struct {
	backends []*ReverseProxyBackend
}

func (this *reverseHash) Pick(req *http.Request) http.Handler {
	if len(this.backends) == 1 {
		return this.backends[0]
	}

	idx := hash(req.RemoteAddr) % len(this.backends)

	return this.backends[idx]
}

func (this *reverseHash) String() string {
	return fmt.Sprintf("[HASH] %d backends with fnv.New64", len(this.backends))
}

func hash(s string) int {
	h := fnv.New64()
	h.Write([]byte(s))
	return int(h.Sum64())
}

func Least(backends []*ReverseProxyBackend) (*reverseLeast, error) {
	if len(backends) == 0 {
		return nil, errBackendsRequired
	}

	return &reverseLeast{
		backends: backends,
	}, nil
}

type reverseLeast struct {
	backends []*ReverseProxyBackend
}

func (this *reverseLeast) Pick(req *http.Request) http.Handler {
	if len(this.backends) == 1 {
		return this.backends[0]
	}

	size := len(this.backends)
	least := this.backends[0].Count
	choice := []*ReverseProxyBackend{
		this.backends[0],
	}

	for i := 1; i < size; i++ {
		b := this.backends[i]
		if b.Count > least {
			continue
		}

		if b.Count == least {
			choice = append(choice, b)
		}

		if b.Count < least {
			least = b.Count
			choice = []*ReverseProxyBackend{
				b,
			}
		}
	}

	if len(choice) == 1 {
		return choice[0]
	}

	return choice[rand.Intn(len(choice))]
}

func (this *reverseLeast) String() string {
	return fmt.Sprintf("[LEAST] %d backends", len(this.backends))
}
