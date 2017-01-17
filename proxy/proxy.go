package proxy

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gobwas/glob"
)

type Proxy struct {
	cfg *proxyConfig
	app *App

	hostPattern []glob.Glob
	uriPattern  []glob.Glob

	reverse *ReverseProxy
	handler http.Handler
}

func (this *Proxy) initialize() {
	if host := this.cfg.Host; host != "" {
		for _, one := range str2NonEmptySlice(host, Sep) {
			log.Printf("[PROXY][%s] host pattern %q added", this.Name(), one)
			this.hostPattern = append(this.hostPattern, glob.MustCompile(one))
		}
	}

	for _, one := range str2NonEmptySlice(this.cfg.URI, Sep) {
		if !strings.HasSuffix(one, Wildcard) {
			one = one + Wildcard
		}

		log.Printf("[PROXY][%s] uri pattern %q added", this.Name(), one)
		this.uriPattern = append(this.uriPattern, glob.MustCompile(one))
	}

	log.Printf("[PROXY][%s] grpc-enabled %v", this.Name(), this.GRPC())
}

func (this *Proxy) build() error {
	reverse, err := buildReverseProxy(this)
	if err != nil {
		return err
	}

	this.reverse = reverse

	m := this.app.server.copyMiddleware()

	var h http.Handler = this.reverse
	for _, one := range m {
		h = one(h)
	}

	this.handler = h
	return nil
}

func (this *Proxy) Match(req *http.Request) bool {
	if !this.matchHost(req) {
		return false
	}

	return this.matchURI(req)
}

func (this *Proxy) matchHost(req *http.Request) bool {
	if len(this.hostPattern) == 0 {
		return true
	}

	for _, pattern := range this.hostPattern {
		if pattern.Match(req.Host) {
			return true
		}
	}

	return false
}

func (this *Proxy) matchURI(req *http.Request) bool {
	for _, pattern := range this.uriPattern {
		if pattern.Match(req.RequestURI) {
			return true
		}
	}

	return false
}

func (this *Proxy) GRPC() bool {
	return this.cfg.GetGRPC()
}

func (this *Proxy) Name() string {
	return fmt.Sprintf("%q-%q", this.app.Name(), this.cfg.Name)
}
