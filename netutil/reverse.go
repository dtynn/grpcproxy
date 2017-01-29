package netutil

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	_ http.Handler = &ReverseProxyBackend{}
	_ Balancer     = &reverseRandom{}
)

func NewReverseProxyBackend(rawBack string, target *url.URL, weight int, transport http.RoundTripper) *ReverseProxyBackend {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport
	return &ReverseProxyBackend{
		Weight:  weight,
		rawBack: rawBack,
		proxy:   proxy,
	}
}

type ReverseProxyBackend struct {
	Weight int

	tls     bool
	rawBack string
	target  *url.URL
	proxy   *httputil.ReverseProxy
}

func (this *ReverseProxyBackend) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Printf("[REVERSE STREAM][%s] %s >>>> %s [W %d]", req.Method, req.URL, this.rawBack, this.Weight)
	this.proxy.ServeHTTP(rw, req)
}

func (this *ReverseProxyBackend) String() string {
	return fmt.Sprintf("%s [W %d]", this.rawBack, this.Weight)
}

func Random(backends []*ReverseProxyBackend) (*reverseRandom, error) {
	if len(backends) == 0 {
		return nil, fmt.Errorf("reverse proxy backends required, got 0")
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

func (this *reverseRandom) Pick() http.Handler {
	rnd := rand.Intn(this.weightN)
	for idx, n := range this.weights {
		if rnd < n {
			return this.backends[idx]
		}
	}

	return this.backends[rand.Intn(len(this.backends))]
}

func (this *reverseRandom) String() string {
	return fmt.Sprintf("%d random reverse backends, weights %v", len(this.backends), this.weights)
}
