package netutil

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"

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

		errorCh: make(chan error, 1),
	}
}

type Server struct {
	h2svr  *http2.Server
	h2opts *http2.ServeConnOpts

	errorCh chan error
}

func (this *Server) Run() error {
	h1svr := this.h2opts.BaseConfig

	l, err := net.Listen("tcp", h1svr.Addr)
	if err != nil {
		return err
	}

	defer l.Close()

	log.Printf("[H2Server][%s] started", h1svr.Addr)

	go this.accept(l)

	err = <-this.errorCh
	log.Printf("[H2Server][%s] stopped, got error %v", h1svr.Addr, err)
	return err
}

func (this *Server) Close() {
	close(this.errorCh)
}

func (this *Server) serve(conn net.Conn) {
	if cfg := this.h2opts.BaseConfig.TLSConfig; cfg != nil {
		tlsConn := tls.Server(conn, cfg)
		if err := tlsConn.Handshake(); err != nil {
			log.Printf("[H2Server][%s] got tls handshake error %s", tlsConn.RemoteAddr(), err)
			conn.Close()
			return
		}

		conn = tlsConn
	}

	this.h2svr.ServeConn(conn, this.h2opts)
}

func (this *Server) accept(l net.Listener) {
	for {
		select {
		case <-this.errorCh:
			return

		default:

		}

		conn, err := l.Accept()
		if err != nil {
			this.errorCh <- err
			return
		}

		go this.serve(conn)
	}
}
