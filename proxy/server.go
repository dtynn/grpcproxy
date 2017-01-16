package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	// "net"
	"net/http"

	// "github.com/facebookgo/grace/gracehttp"
	"github.com/dtynn/grpcproxy/h2grace/h2gracehttp"
	"golang.org/x/net/http2"
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

	servers := make([]*h2gracehttp.Server, 0, len(this.bindMap))

	for bind, apps := range this.bindMap {
		srv := &http.Server{}
		srv.Addr = bind
		srv.Handler = newProxyHandler(apps)

		if len(this.cert) > 0 {
			srv.TLSConfig = &tls.Config{
				Certificates: this.cert,
				NextProtos:   []string{http2.NextProtoTLS},
			}
		}

		http2.ConfigureServer(srv, nil)

		h2Server := &http2.Server{}
		h2SvrOpt := &http2.ServeConnOpts{
			BaseConfig: srv,
		}

		servers = append(servers, h2gracehttp.NewServer(h2Server, h2SvrOpt))

		// go func(lis net.Listener, srv *http2.Server, opts *http2.ServeConnOpts) {
		// 	conn, err := lis.Accept()
		// 	if err != nil {
		// 		log.Fatalln(err)
		// 	}

		// 	srv.ServeConn(conn, opts)
		// }(lis, h2Server, h2SvrOpt)
	}

	h := h2gracehttp.NewHTTP2()
	return h2gracehttp.Serve(h, true, servers...)
}
