package proxy

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"sync"

	"golang.org/x/net/http2"
)

type listener struct {
	bind string
	app  []*App
	srv  *Server

	stop chan bool
}

func (this *listener) run(wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	lis, err := net.Listen("tcp", this.bind)

	if err != nil {
		errCh <- err
		return
	}

	if len(this.srv.cert) != 0 {
		lis = tls.NewListener(lis, &tls.Config{
			Certificates: this.srv.cert,
			NextProtos:   []string{"h2"},
		})
	}

	defer lis.Close()

	log.Printf("[LISTENER] %d apps listen on %q with TLS %v", len(this.app), this.bind, len(this.srv.cert) != 0)

	srv := &http2.Server{}
	opts := &http2.ServeConnOpts{
		Handler: this,
	}

	for {
		select {
		case <-this.stop:
			break

		default:

		}

		conn, err := lis.Accept()
		if err != nil {
			errCh <- err
			return
		}

		go srv.ServeConn(conn, opts)
	}

	log.Printf("service on %s stopped", this.bind)
	return
}

func (this *listener) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, app := range this.app {
		proxy, ok := app.Match(req)
		if ok {
			proxy.handler.ServeHTTP(rw, req)
			return
		}
	}

	log.Printf("[NOT FOUND][%s] %s%s", req.Method, req.Host, req.RequestURI)
	return
}
