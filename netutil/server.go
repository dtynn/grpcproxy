package netutil

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/cockroachdb/cmux"
	"golang.org/x/net/http2"
)

func NewServer(svr *http.Server) *Server {
	return &Server{
		h2opts: &http2.ServeConnOpts{
			BaseConfig: svr,
		},

		errorCh: make(chan error, 3),
		closeCh: make(chan struct{}, 1),
	}
}

type Server struct {
	http2.Server
	h2opts *http2.ServeConnOpts

	mu sync.RWMutex

	errorCh chan error
	closeCh chan struct{}
}

func (this *Server) Run() error {
	bind := this.h2opts.BaseConfig.Addr

	l, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	defer l.Close()

	log.Printf("[H2Server][%s] started", bind)

	go this.mux(l)

	select {
	case err = <-this.errorCh:
		log.Printf("[H2Server][%s] stopped, got serve error %v", bind, err)
	case <-this.closeCh:
		log.Printf("[H2Server][%s] manually stopped", bind)
	}

	return err
}

func (this *Server) Close() {
	close(this.closeCh)
}

func (this *Server) serve(conn net.Conn, isTLS bool) {
	this.mu.RLock()
	tlsCfg := this.h2opts.BaseConfig.TLSConfig
	this.mu.RUnlock()

	if isTLS && tlsCfg != nil {
		tlsConn := tls.Server(conn, tlsCfg)
		if err := tlsConn.Handshake(); err != nil {
			log.Printf("[H2Server][%s] got tls handshake error %s", tlsConn.RemoteAddr(), err)
			conn.Close()
			return
		}

		conn = tlsConn
	}

	this.ServeConn(conn, this.h2opts)
}

func (this *Server) Reload(tlsCfg *tls.Config, handler http.Handler) {
	this.mu.Lock()
	this.h2opts.BaseConfig.TLSConfig = tlsCfg
	this.h2opts.Handler = handler
	this.mu.Unlock()
}

func (this *Server) accept(l net.Listener, isTLS bool) {
	for {
		select {
		case <-this.closeCh:
			return

		default:

		}

		conn, err := l.Accept()

		if err != nil {
			this.errorCh <- err
			return
		}

		go this.serve(conn, isTLS)
	}
}

func (this *Server) mux(l net.Listener) {
	m := cmux.New(l)

	lH2 := m.Match(cmux.HTTP2())
	defer lH2.Close()

	lTLS := m.Match(cmux.Any())
	defer lTLS.Close()

	go this.accept(lH2, false)
	go this.accept(lTLS, true)

	this.errorCh <- m.Serve()
}
