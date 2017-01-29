package netutil

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	_ http.Handler = &ReverseProxyBackend{}
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

	Count   int64
	rawBack string
	target  *url.URL
	proxy   *httputil.ReverseProxy
}

func (this *ReverseProxyBackend) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	this.Count += 1
	log.Printf("[REVERSE STREAM][%s] %s >>>> %s [W %d]", req.Method, req.URL, this.rawBack, this.Weight)
	this.proxy.ServeHTTP(rw, req)
}

func (this *ReverseProxyBackend) String() string {
	return fmt.Sprintf("%s [W %d]", this.rawBack, this.Weight)
}
