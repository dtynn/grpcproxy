package proxy

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

func newListener(svr *http.Server, errorCh chan<- error) *listener {
	return &listener{
		h2svr: &http2.Server{},
		h2opt: &http2.ServeConnOpts{
			BaseConfig: svr,
		},
		closeCh: make(chan struct{}, 1),
		errorCh: errorCh,
	}
}

type listener struct {
	h2svr *http2.Server
	h2opt *http2.ServeConnOpts

	closeCh chan struct{}
	errorCh chan<- error
}

func (this *listener) run() {
	cfg := this.h2opt.BaseConfig
	l, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		this.errorCh <- err
		return
	}

	defer l.Close()

	log.Printf("[LISTENER] listen on %s", cfg.Addr)

	for {
		select {
		case <-this.closeCh:
			break

		default:

		}

		conn, err := l.Accept()
		if err != nil {
			this.errorCh <- err
			break
		}

		go this.serve(conn)
	}

	log.Printf("[LISTENER] stop on %s", cfg.Addr)
	return
}

func (this *listener) serve(conn net.Conn) {
	if this.h2opt.BaseConfig.TLSConfig != nil {
		tlsConn := tls.Server(conn, this.h2opt.BaseConfig.TLSConfig)
		// TODO: set read deadline?
		if err := tlsConn.Handshake(); err != nil {
			log.Printf("[LISTENER] got handshake error for %s : %s", tlsConn.RemoteAddr(), err)
			conn.Close()
			return
		}

		conn = tlsConn
	}

	this.h2svr.ServeConn(conn, this.h2opt)
}

func (this *listener) close() {
	close(this.closeCh)
}
