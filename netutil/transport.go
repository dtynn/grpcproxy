package netutil

import (
	"crypto/tls"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

var (
	_ http.RoundTripper = &Transport{}
)

type TransportOpt struct {
	Trailer         []string
	AllowHTTP       bool
	TLSClientConfig *tls.Config
}

func NewTransport(opt TransportOpt) *Transport {
	h2t := &http2.Transport{
		AllowHTTP:       opt.AllowHTTP,
		TLSClientConfig: opt.TLSClientConfig,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			conn, err := net.Dial(network, addr)
			if err != nil {
				return nil, err
			}

			if opt.TLSClientConfig != nil {
				conn = tls.Client(conn, cfg)
			}

			return conn, nil
		},
	}

	return &Transport{
		opt: opt,
		h2t: h2t,
	}
}

type Transport struct {
	opt TransportOpt
	h2t *http2.Transport
}

func (this *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := this.h2t.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if len(this.opt.Trailer) > 0 {
		if resp.Trailer == nil {
			resp.Trailer = http.Header{}
		}

		for _, key := range this.opt.Trailer {
			value := resp.Header.Get(key)
			resp.Trailer.Set(key, value)
		}
	}

	return resp, nil
}
