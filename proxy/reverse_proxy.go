package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/http2"
)

const (
	grpcStatusHeader  = "Grpc-Status"
	grpcMessageHeader = "Grpc-Message"
)

type ReverseProxyBackend struct {
	backUrl string
	target  *url.URL
	proxy   *httputil.ReverseProxy
	weight  int
}

type ReverseProxy struct {
	policy    string
	backends  []*ReverseProxyBackend
	transport *proxyTransport
}

func buildReverseProxy(p *Proxy) (*ReverseProxy, error) {
	tlsOn := p.cfg.TLS

	t2 := &http2.Transport{
		AllowHTTP: !tlsOn,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: p.cfg.InsecureSkipVerify,
		},
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			conn, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}

			if tlsOn {
				conn = tls.Client(conn, cfg)
			}
			return conn, nil
		},
	}

	// reverse proxy
	revr := &ReverseProxy{
		policy: p.cfg.Policy,
		transport: &proxyTransport{
			grpc:      p.GRPC(),
			transport: t2,
		},
	}

	// build backends
	backends := make([]*ReverseProxyBackend, 0)

	for _, back := range str2NonEmptySlice(p.cfg.Backend, Sep) {
		weight := 1

		if pieces := str2NonEmptySlice(back, ";"); len(pieces) == 2 {
			back = pieces[0]
			if w, _ := strconv.Atoi(strings.TrimSpace(pieces[1])); w != 0 {
				weight = w
			}
		}

		back = strings.TrimSpace(back)
		if back == "" {
			continue
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
			if tlsOn {
				target.Scheme = "https"
			} else {
				target.Scheme = "http"
			}
		}

		rvProxy := httputil.NewSingleHostReverseProxy(target)
		rvProxy.Transport = revr.transport

		backends = append(backends, &ReverseProxyBackend{
			backUrl: back,
			target:  target,
			proxy:   rvProxy,
			weight:  weight,
		})
	}

	if len(backends) == 0 {
		return nil, fmt.Errorf("[BACKEND] backends required", p.Name())
	}

	for _, one := range backends {
		log.Printf("[BACKEND][%s] %s weight %d added", p.Name(), one.backUrl, one.weight)
	}

	revr.backends = backends

	return revr, nil
}

func (this *ReverseProxy) pickBackend() *ReverseProxyBackend {
	if len(this.backends) == 1 {
		return this.backends[0]
	}

	idx := rand.Intn(len(this.backends))
	return this.backends[idx]
}

func (this *ReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	backend := this.pickBackend()

	log.Printf("[STREAM] [%s%s] => [%s]", req.Host, req.RequestURI, backend.backUrl)
	backend.proxy.ServeHTTP(rw, req)
}

type proxyTransport struct {
	grpc      bool
	transport http.RoundTripper
}

func (this *proxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := this.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if this.grpc && req.Header.Get("Content-Type") == "application/grpc" {
		status := resp.Header.Get(grpcStatusHeader)
		message := resp.Header.Get(grpcMessageHeader)

		if status == "" {
			status = "0"
		}

		resp.Trailer = http.Header{}
		resp.Trailer.Set(grpcStatusHeader, status)
		resp.Trailer.Set(grpcMessageHeader, message)
	}

	return resp, err
}
