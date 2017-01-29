package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dtynn/grpcproxy/config"
	"github.com/dtynn/grpcproxy/netutil"
	"github.com/gobwas/glob"
)

var (
	grpcTrailerHeaders = []string{
		"Grpc-Status",
		"Grpc-Message",
	}
)

func NewProxy(app *App, cfg *config.ProxyConfig) (*Proxy, error) {
	proxy := &Proxy{
		app: app,
		cfg: cfg,
	}

	// host patterns
	if host := cfg.Host; host != "" {
		for _, one := range str2NonEmptySlice(host, Sep) {
			log.Printf("[PROXY][%s] host pattern %q added", proxy, one)
			proxy.hosts = append(proxy.hosts, glob.MustCompile(one))
		}
	}

	// uri patterns
	for _, one := range str2NonEmptySlice(cfg.URI, Sep) {
		if !strings.HasSuffix(one, Wildcard) {
			one = one + Wildcard
		}

		log.Printf("[PROXY][%s] uri pattern %q added", proxy, one)
		proxy.uris = append(proxy.uris, glob.MustCompile(one))
	}

	// http2 transport
	h2topt := netutil.TransportOpt{
		AllowHTTP: !cfg.TLS,
	}

	if cfg.TLS {
		h2topt.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		}
	}

	grpcEnabled := cfg.GetGRPC()
	log.Printf("[PROXY][%s] grpc enabled %v", proxy, grpcEnabled)

	if grpcEnabled {
		h2topt.Trailer = grpcTrailerHeaders
	}

	h2t := netutil.NewTransport(h2topt)

	// reverse proxy backends
	backends := make([]*netutil.ReverseProxyBackend, 0)
	for _, back := range str2NonEmptySlice(cfg.Backend, Sep) {
		weight := 1

		if pieces := str2NonEmptySlice(back, ";"); len(pieces) == 2 {
			back = pieces[0]
			if w, _ := strconv.Atoi(strings.TrimSpace(pieces[1])); w > 0 {
				weight = w
			}
		}

		target, err := buildTargetUrl(cfg.TLS, back)
		if err != nil {
			return nil, err
		}

		if target == nil {
			continue
		}

		backend := netutil.NewReverseProxyBackend(back, target, weight, h2t)
		log.Printf("[PROXY][%s] backend %q added", proxy, backend)

		backends = append(backends, backend)
	}

	var balancer netutil.Balancer

	switch cfg.Policy {
	default:
		b, err := netutil.Random(backends)
		if err != nil {
			return nil, err
		}

		balancer = b
	}

	log.Printf("[PROXY][%s] use balancer %q", proxy, balancer)

	proxy.balancer = balancer

	return proxy, nil
}

type Proxy struct {
	app *App
	cfg *config.ProxyConfig

	hosts []glob.Glob
	uris  []glob.Glob

	balancer netutil.Balancer
}

func (this *Proxy) Match(req *http.Request) bool {
	if !this.matchHost(req) {
		return false
	}

	return this.matchURI(req)
}

func (this *Proxy) matchHost(req *http.Request) bool {
	if len(this.hosts) == 0 {
		return true
	}

	for _, pattern := range this.hosts {
		if pattern.Match(req.Host) {
			return true
		}
	}

	return false
}

func (this *Proxy) matchURI(req *http.Request) bool {
	for _, pattern := range this.uris {
		if pattern.Match(req.RequestURI) {
			return true
		}
	}

	return false
}

func (this *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h := this.balancer.Pick()
	h.ServeHTTP(rw, req)
}

func (this *Proxy) String() string {
	return fmt.Sprintf("%s-%s", this.app, this.cfg.Name)
}

func buildTargetUrl(tls bool, back string) (*url.URL, error) {
	back = strings.TrimSpace(back)
	if back == "" {
		return nil, nil
	}

	target, err := url.Parse(back)
	if err != nil {
		return nil, err
	}

	if target.Host == "" && target.Scheme != "" && target.Opaque != "" {
		// like "dev.yogurbox.com:51000", "localhost:51000"

		target.Host = fmt.Sprintf("%s:%s", target.Scheme, target.Opaque)
		target.Opaque = ""
		target.Scheme = ""

	} else if target.Host == "" && target.Path != "" {
		// like "127.0.0.1:51000"

		target.Host = target.Path
		target.Path = ""
	}

	if target.Scheme == "" {
		if tls {
			target.Scheme = "https"
		} else {
			target.Scheme = "http"
		}
	}

	return target, nil
}
