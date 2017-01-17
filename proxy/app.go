package proxy

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gobwas/glob"
)

type App struct {
	cfg    *appConfig
	server *Server

	hostPattern []glob.Glob

	Proxy []*Proxy
}

func (this *App) initialize() {
	// init host
	host := this.cfg.Host

	for _, one := range str2NonEmptySlice(host, Sep) {
		this.hostPattern = append(this.hostPattern, glob.MustCompile(one))
		log.Printf("[APP][%s] host pattern %q added", this.Name(), one)
	}

	log.Printf("[APP][%q] bind on %s", this.Name(), this.cfg.GetBind())
	log.Printf("[APP][%q] grpc-enabled %v", this.Name(), this.cfg.GetGRPC())

	for _, proxyCfg := range this.cfg.Proxy {
		proxy := &Proxy{
			cfg: proxyCfg,
			app: this,
		}

		proxy.initialize()

		this.Proxy = append(this.Proxy, proxy)
	}
}

func (this *App) build() error {
	for _, proxy := range this.Proxy {
		if err := proxy.build(); err != nil {
			return fmt.Errorf("[PROXY BUILD ERROR][%s] %s", proxy.Name(), err)
		}
	}

	return nil
}

func (this *App) Match(req *http.Request) (*Proxy, bool) {
	if !this.matchHost(req) {
		return nil, false
	}

	for _, proxy := range this.Proxy {
		if proxy.Match(req) {
			return proxy, true
		}
	}

	return nil, false
}

func (this *App) matchHost(req *http.Request) bool {
	for _, pattern := range this.hostPattern {
		if pattern.Match(req.Host) {
			return true
		}
	}

	return false
}

func (this *App) Bind() []string {
	return str2NonEmptySlice(this.cfg.GetBind(), Sep)
}

func (this *App) Name() string {
	return this.cfg.Name
}
