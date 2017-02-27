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
		closeCh: make(chan struct{}, 1),
	}
}

type Service struct {
	cfgFilePath string
	initialized bool

	cfg config.ServerConfig

	svrs []*netutil.Server

	closeCh chan struct{}
	mu      sync.RWMutex
}

func (this *Service) Init(cfg config.ServerConfig) error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.initialized {
		return fmt.Errorf("[SERVER] already initialized")
	}

	apps, err := this.buildApps(&cfg)
	if err != nil {
		return err
	}

	cert, err := loadCerts(cfg.Cert)
	if err != nil {
		return err
	}

	bindings := nonEmptySlice(cfg.Bind)
	if len(bindings) == 0 {
		return fmt.Errorf("[SERVER] bindings required")
	}

	log.Printf("[SERVER] bind on %v", bindings)

	svrs := make([]*netutil.Server, 0, len(bindings))
	for _, bind := range bindings {
		svr := &http.Server{}
		svr.Addr = bind
		svr.Handler = apps
		if len(cert) > 0 {
			svr.TLSConfig = &tls.Config{
				Certificates: cert,
				NextProtos:   netutil.NextProtos,
			}
		}
		svrs = append(svrs, netutil.NewServer(svr))
	}

	this.svrs = svrs
	return nil
}

func (this *Service) ReloadConfigFile() error {
	cfg, err := config.ReadConfig(this.cfgFilePath)
	if err != nil {
		return err
	}

	return this.Reload(cfg)
}

func (this *Service) Reload(cfg config.ServerConfig) error {
	log.Printf("[SERVER] reloading")
	// init apps
	apps, err := this.buildApps(&cfg)
	if err != nil {
		return err
	}

	cert, err := loadCerts(cfg.Bind)
	if err != nil {
		return err
	}

	var tlsCfg *tls.Config
	if len(cert) > 0 {
		tlsCfg = &tls.Config{
			Certificates: cert,
			NextProtos:   netutil.NextProtos,
		}
	}

	for _, svr := range this.svrs {
		svr.Reload(tlsCfg, apps)
	}

	return nil
}

func (this *Service) Run() error {
	this.mu.Lock()
	initialized := this.initialized
	this.mu.Unlock()

	if !initialized {
		return fmt.Errorf("server not initialized")
	}

	servers := this.svrs

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
	close(this.closeCh)
}

func (this *Service) buildApps(cfg *config.ServerConfig) (Apps, error) {
	apps := make(Apps, 0, len(cfg.App))

	for _, appCfg := range cfg.App {
		app, err := NewApp(this, appCfg)
		if err != nil {
			return nil, err
		}

		apps = append(apps, app)
	}

	return apps, nil
}

func loadCerts(path []string) ([]tls.Certificate, error) {
	certs := []tls.Certificate{}
	if len(path) == 2 {
		cert, err := tls.LoadX509KeyPair(path[0], path[1])
		if err != nil {
			return nil, err
		}

		certs = append(certs, cert)
	}

	return certs, nil
}
