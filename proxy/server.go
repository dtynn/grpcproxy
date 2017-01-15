package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Middleware func(handler http.Handler) http.Handler

func NewServer(cfg ServerConfig) *Server {
	srv := &Server{
		cfg: cfg,
	}

	srv.initialize()

	return srv
}

type Server struct {
	cfg ServerConfig

	App        []*App
	middleware []Middleware

	cert    []tls.Certificate
	bindMap map[string][]*App

	closeCh chan bool
}

func (this *Server) initialize() {
	log.Printf("[VERSION] %s", version.String())
	log.Println("[SERVER INITAILIZE] start")
	this.closeCh = make(chan bool, 1)

	for _, appCfg := range this.cfg.App {
		app := &App{
			cfg:    appCfg,
			server: this,
		}

		app.initialize()

		this.App = append(this.App, app)
	}
	log.Println("[SERVER INITAILIZE] finished")
}

func (this *Server) build() error {
	log.Println("[SERVER BUILD] start")

	if len(this.cfg.TLS) == 2 {
		cert, err := tls.LoadX509KeyPair(this.cfg.TLS[0], this.cfg.TLS[1])
		if err != nil {
			return fmt.Errorf("[SERVER BUILD ERROR] load cert files: %s", err)
		}

		this.cert = []tls.Certificate{cert}
		log.Printf("[SERVER BUILD] TLS %s loaded", this.cfg.TLS)
	}

	bindMap := map[string][]*App{}

	for _, app := range this.App {
		if err := app.build(); err != nil {
			return fmt.Errorf("[APP BUILD ERROR] %s", err)
		}

		bindMap[app.Bind()] = append(bindMap[app.Bind()], app)
	}

	this.bindMap = bindMap
	log.Println("[SERVER BUILD] finished")
	return nil
}

func (this *Server) copyMiddleware() []Middleware {
	res := make([]Middleware, len(this.middleware))
	for i, m := range this.middleware {
		res[i] = m
	}

	return res
}

func (this *Server) Use(m ...Middleware) {
	this.middleware = append(this.middleware, m...)
}

func (this *Server) Run() error {
	if err := this.build(); err != nil {
		return err
	}

	listeners := make([]*listener, 0, len(this.bindMap))

	for bind, app := range this.bindMap {
		stopCh := make(chan bool, 1)
		lis := &listener{
			bind: bind,
			app:  app,
			srv:  this,
			stop: stopCh,
		}

		listeners = append(listeners, lis)
	}

	errCh := make(chan error)
	var wg sync.WaitGroup

	for _, lis := range listeners {
		wg.Add(1)

		go func(lis *listener, wg *sync.WaitGroup, errCh chan<- error) {
			lis.run(wg, errCh)
		}(lis, &wg, errCh)
	}

	var err error

	for {
		select {
		case err = <-errCh:
			break

		case <-this.closeCh:
			break
		}
	}

	for _, lis := range listeners {
		close(lis.stop)
	}

	wg.Wait()

	return err
}

func (this *Server) Close() {
	close(this.closeCh)
}
