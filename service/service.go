package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/dtynn/grpcproxy/config"
	"github.com/dtynn/grpcproxy/netutil"
)

func NewServiceWithCfgFile(cfgFilePath string) (*Service, error) {
	service := NewService()
	service.cfgFilePath = cfgFilePath

	err := service.ReloadConfigFile()
	if err != nil {
		return nil, err
	}

	return service, nil
}

func NewService() *Service {
	return &Service{
		stopCh:  make(chan struct{}, 1),
		closeCh: make(chan struct{}, 1),
	}
}

type Service struct {
	cfgFilePath string

	cfg      config.ServerConfig
	apps     []*App
	cert     []tls.Certificate
	bindings map[string]Apps

	closeCh chan struct{}
	stopCh  chan struct{}
	mu      sync.Mutex
}

func (this *Service) ReloadConfigFile() error {
	cfg, err := config.ReadConfig(this.cfgFilePath)
	if err != nil {
		return err
	}

	return this.Load(cfg)
}

func (this *Service) Load(cfg config.ServerConfig) error {
	// init apps
	bindings := map[string]Apps{}
	apps := []*App{}

	for _, appCfg := range cfg.App {
		app, err := NewApp(this, appCfg)
		if err != nil {
			return err
		}

		apps = append(apps, app)

		for _, bind := range app.Bind() {
			bindings[bind] = append(bindings[bind], app)
		}
	}

	certs := []tls.Certificate{}

	// load cert files
	if len(cfg.Cert) == 2 {
		cert, err := tls.LoadX509KeyPair(cfg.Cert[0], cfg.Cert[1])
		if err != nil {
			return fmt.Errorf("[SERVER LOAD] fail to load cert files %v: %s", cfg.Cert, err)
		}

		certs = append(certs, cert)
		log.Printf("[SERVER LOAD] cert files %s loaded", cfg.Cert)
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	this.cfg = cfg
	this.apps = apps
	this.cert = certs
	this.bindings = bindings

	return nil
}

func (this *Service) Run() error {
	this.mu.Lock()
	defer this.mu.Unlock()

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
	this.closeCh <- struct{}{}
}

func (this *Service) Stop() {
	this.Close()
	close(this.stopCh)
}

func (this *Service) Wait() {
	<-this.stopCh
}
