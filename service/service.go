package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dtynn/grpcproxy/config"
	"github.com/dtynn/grpcproxy/netutil"
)

func NewService(cfg config.ServerConfig) (*Service, error) {
	service := &Service{
		cfg:     cfg,
		closeCh: make(chan struct{}, 1),
	}

	// init apps
	service.bindings = map[string]Apps{}

	for _, appCfg := range cfg.App {
		app, err := NewApp(service, appCfg)
		if err != nil {
			return nil, err
		}

		service.apps = append(service.apps, app)

		for _, bind := range app.Bind() {
			service.bindings[bind] = append(service.bindings[bind], app)
		}
	}

	// load cert files
	if len(cfg.Cert) == 2 {
		cert, err := tls.LoadX509KeyPair(cfg.Cert[0], cfg.Cert[1])
		if err != nil {
			return nil, fmt.Errorf("[SERVER BUILD] fail to load cert files %v: %s", cfg.Cert, err)
		}

		service.cert = []tls.Certificate{cert}
		log.Printf("[SERVER BUILD] cert files %s loaded", cfg.Cert)
	}

	return service, nil
}

type Service struct {
	cfg      config.ServerConfig
	apps     []*App
	cert     []tls.Certificate
	bindings map[string]Apps

	closeCh chan struct{}
}

func (this *Service) Run() error {
	servers := make([]*netutil.Server, 0, len(this.bindings))
	for bind, apps := range this.bindings {
		svr := &http.Server{}
		svr.Addr = bind
		svr.Handler = apps
		if len(this.cert) > 0 {
			svr.TLSConfig = &tls.Config{
				Certificates: this.cert,
				NextProtos:   netutil.NextProtos,
			}
		}

		server := netutil.NewServer(svr, nil)
		servers = append(servers, server)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(servers))

	for _, server := range servers {
		wg.Add(1)
		go func(server *netutil.Server, wg *sync.WaitGroup, errCh chan error) {
			if err := server.Run(); err != nil {
				errCh <- err
			}

			wg.Done()
		}(server, &wg, errCh)
	}

	go this.signalHandler()

	var err error

	select {
	case err = <-errCh:

	case <-this.closeCh:

	}

	for _, server := range servers {
		server.Close()
	}

	wg.Wait()

	return err
}

func (this *Service) Close() {
	close(this.closeCh)
}

func (this *Service) signalHandler() {
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
