package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/net/http2"
)

type Middleware func(handler http.Handler) http.Handler

func NewServer(cfg ServerConfig) *Server {
	svr := &Server{
		cfg:     cfg,
		closeCh: make(chan struct{}, 1),
	}

	svr.initialize()

	return svr
}

type Server struct {
	cfg ServerConfig

	App        []*App
	middleware []Middleware

	cert    []tls.Certificate
	bindMap map[string][]*App

	closeCh chan struct{}
}

func (this *Server) initialize() {
	log.Printf("[VERSION] %s", version.String())
	log.Println("[SERVER INITAILIZE] start")

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

	if len(this.cfg.Cert) == 2 {
		cert, err := tls.LoadX509KeyPair(this.cfg.Cert[0], this.cfg.Cert[1])
		if err != nil {
			return fmt.Errorf("[SERVER BUILD ERROR] load cert files: %s", err)
		}

		this.cert = []tls.Certificate{cert}
		log.Printf("[SERVER BUILD] cert %s loaded", this.cfg.Cert)
	}

	bindMap := map[string][]*App{}

	for _, app := range this.App {
		if err := app.build(); err != nil {
			return fmt.Errorf("[APP BUILD ERROR] %s", err)
		}

		for _, bind := range app.Bind() {
			bindMap[bind] = append(bindMap[bind], app)
		}
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
	errCh := make(chan error, len(this.bindMap))
	var wg sync.WaitGroup

	for bind, apps := range this.bindMap {
		svr := &http.Server{}
		svr.Addr = bind
		svr.Handler = newProxyHandler(apps)
		if len(this.cert) > 0 {
			svr.TLSConfig = &tls.Config{
				Certificates: this.cert,
				NextProtos:   []string{http2.NextProtoTLS},
			}
		}

		wg.Add(1)
		listener := newListener(svr, &wg, errCh)
		listeners = append(listeners, listener)
		go listener.run()
	}

	go this.signalHandler()
	var err error

	select {
	case err = <-errCh:
		return err

	case <-this.closeCh:

	}

	for i, _ := range listeners {
		listeners[i].close()
	}

	wg.Wait()

	log.Printf("[SERVER] shutdown")
	return err
}

func (this *Server) Close() {
	close(this.closeCh)
}

func (this *Server) signalHandler() {
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			// this ensures a subsequent INT/TERM will trigger standard go behaviour of
			// terminating.
			log.Printf("[SERVER] got signal %s", sig)
			signal.Stop(ch)
			this.Close()
			return
		}
	}
}
