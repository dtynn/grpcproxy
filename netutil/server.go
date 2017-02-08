package netutil

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"

	"github.com/cockroachdb/cmux"
	"golang.org/x/net/http2"
)

func NewServer(svr *http.Server, h2svr *http2.Server) *Server {
	if h2svr == nil {
		h2svr = &http2.Server{}
	}

	return &Server{
		h2svr: h2svr,
		h2opts: &http2.ServeConnOpts{
			BaseConfig: svr,
		},

		errorCh: make(chan error, 3),
		closeCh: make(chan struct{}, 1),
	}
}

type Server struct {
	h2svr  *http2.Server
	h2opts *http2.ServeConnOpts

	errorCh chan error
	closeCh chan struct{}
}

func (this *Server) Run() error {
	h1svr := this.h2opts.BaseConfig

	l, err := net.Listen("tcp", h1svr.Addr)
	if err != nil {
		return err
	}

	defer l.Close()

	log.Printf("[H2Server][%s] started", h1svr.Addr)

	go this.mux(l)

	select {
	case err = <-this.errorCh:
		log.Printf("[H2Server][%s] stopped, got serve error %v", h1svr.Addr, err)
	case <-this.closeCh:
		log.Printf("[H2Server][%s] manually stopped", h1svr.Addr)
	}

	return err
}

func (this *Server) Close() {
	close(this.closeCh)
}

func (this *Server) serve(conn net.Conn, isTLS bool) {
	if isTLS && this.h2opts.BaseConfig.TLSConfig != nil {
		tlsConn := tls.Server(conn, this.h2opts.BaseConfig.TLSConfig)
		if err := tlsConn.Handshake(); err != nil {
			log.Printf("[H2Server][%s] got tls handshake error %s", tlsConn.RemoteAddr(), err)
			conn.Close()
			return
		}

		conn = tlsConn
	}

	this.h2svr.ServeConn(conn, this.h2opts)
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
