package service

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dtynn/grpcproxy/config"
	"github.com/gobwas/glob"
)

func NewApp(service *Service, cfg *config.AppConfig) (*App, error) {
	app := &App{
		service: service,
		cfg:     cfg,
	}

	host := cfg.Host

	for _, one := range str2NonEmptySlice(host, Sep) {
		log.Printf("[APP][%s] host pattern %q added", app, one)
		app.hosts = append(app.hosts, glob.MustCompile(one))
	}

	for _, proxyCfg := range cfg.Proxy {
		proxy, err := NewProxy(app, proxyCfg)
		if err != nil {
			return nil, fmt.Errorf("[APP][%s] got proxy init error %q", app, err)
		}

		app.Proxy = append(app.Proxy, proxy)
	}

	return app, nil
}

type App struct {
	service *Service
	cfg     *config.AppConfig

	hosts []glob.Glob

	Proxy []*Proxy
}

func (this *App) String() string {
	return this.cfg.Name
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
	for _, pattern := range this.hosts {
		if pattern.Match(req.Host) {
			return true
		}
	}

	return false
}

type Apps []*App

func (this Apps) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, app := range this {
		if proxy, ok := app.Match(req); ok {
			proxy.ServeHTTP(rw, req)
			return
		}
	}

	log.Printf("[NOT FOUND][%s] %s%s", req.Method, req.Host, req.RequestURI)
}
