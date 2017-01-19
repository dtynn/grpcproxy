package proxy

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"

	"golang.org/x/net/http2"
)

func newListener(svr *http.Server, wg *sync.WaitGroup, errorCh chan<- error) *listener {
	return &listener{
		h2svr: &http2.Server{},
		h2opt: &http2.ServeConnOpts{
			BaseConfig: svr,
		},
		wg:      wg,
		closeCh: make(chan struct{}, 1),
		errorCh: errorCh,
		connCh:  make(chan net.Conn, 1),
	}
}

type listener struct {
	h2svr *http2.Server
	h2opt *http2.ServeConnOpts

	wg *sync.WaitGroup

	closeCh chan struct{}
	errorCh chan<- error
	connCh  chan net.Conn
}

func (this *listener) run() {
	defer this.wg.Done()

	cfg := this.h2opt.BaseConfig
	l, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		this.errorCh <- err
		return
	}

	defer l.Close()

	log.Printf("[LISTENER] listen on %s", cfg.Addr)
	go this.accept(l)

LOOP:
	for {
		select {
		case <-this.closeCh:
			break LOOP

		case conn := <-this.connCh:
			go this.serve(conn)
		}

	}

	log.Printf("[LISTENER] stop on %s", cfg.Addr)
	return
}

func (this *listener) accept(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			this.errorCh <- err
			return
		}

		this.connCh <- conn
	}
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
