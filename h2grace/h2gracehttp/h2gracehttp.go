// fork version for github.com/facebookgo/grace/gracehttp

package h2gracehttp

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/dtynn/grpcproxy/h2grace/http2down"
	// "github.com/facebookgo/grace/gracenet"
	"golang.org/x/net/http2"
)

var (
	didInherit = os.Getenv("LISTEN_FDS") != ""
	ppid       = os.Getppid()
)

func NewHTTP2() *http2down.HTTP2 {
	return &http2down.HTTP2{}
}

func NewServer(s *http2.Server, opts *http2.ServeConnOpts) *Server {
	return &Server{
		s:    s,
		opts: opts,
	}
}

type Server struct {
	s    *http2.Server
	opts *http2.ServeConnOpts
}

func Serve(h *http2down.HTTP2, verbose bool, servers ...*Server) error {
	a := newApp(servers)
	if h == nil {
		h = &http2down.HTTP2{}
	}
	a.hdown = h

	// Acquire Listeners
	if err := a.listen(); err != nil {
		return err
	}

	// Some useful logging.
	if verbose {
		if didInherit {
			if ppid == 1 {
				log.Printf("[H2GRACEHTTP]Listening on init activated %s", pprintAddr(a.listeners))
			} else {
				const msg = "[H2GRACEHTTP]Graceful handoff of %s with new pid %d and old pid %d"
				log.Printf(msg, pprintAddr(a.listeners), os.Getpid(), ppid)
			}
		} else {
			const msg = "[H2GRACEHTTP]Serving %s with pid %d"
			log.Printf(msg, pprintAddr(a.listeners), os.Getpid())
		}
	}

	// Start serving.
	a.serve()

	// Close the parent if we inherited and it wasn't init that started us.
	if didInherit && ppid != 1 {
		if err := syscall.Kill(ppid, syscall.SIGTERM); err != nil {
			return fmt.Errorf("failed to close parent: %s", err)
		}
	}

	waitdone := make(chan struct{})
	go func() {
		defer close(waitdone)
		a.wait()
	}()

	select {
	case err := <-a.errors:
		if err == nil {
			panic("unexpected nil error")
		}
		return err
	case <-waitdone:
		if verbose {
			log.Printf("[H2GRACEHTTP]Exiting pid %d.", os.Getpid())
		}
		return nil
	}
}

func newApp(servers []*Server) *app {
	return &app{
		servers:   servers,
		hdown:     &http2down.HTTP2{},
		listeners: make([]net.Listener, 0, len(servers)),
		sds:       make([]http2down.Server, 0, len(servers)),
		errors:    make(chan error, (len(servers)*2)+1),
	}
}

type app struct {
	servers   []*Server
	hdown     *http2down.HTTP2
	listeners []net.Listener
	sds       []http2down.Server
	errors    chan error
}

func (this *app) listen() error {
	for _, s := range this.servers {
		// As for http2 server, we would not use tlsConfig
		l, err := net.Listen("tcp", s.opts.BaseConfig.Addr)
		if err != nil {
			return err
		}

		this.listeners = append(this.listeners, l)
	}
	return nil
}

func (this *app) serve() {
	for i, s := range this.servers {
		this.sds = append(this.sds, this.hdown.Serve(this.listeners[i], s.s, s.opts))
	}
}

func (this *app) wait() {
	var wg sync.WaitGroup
	wg.Add(len(this.sds) * 2) // Wait & Stop
	go this.signalHandler(&wg)
	for _, s := range this.sds {
		go func(s http2down.Server) {
			defer wg.Done()
			if err := s.Wait(); err != nil {
				this.errors <- err
			}
		}(s)
	}
	wg.Wait()
}

func (this *app) term(wg *sync.WaitGroup) {
	for _, s := range this.sds {
		go func(s http2down.Server) {
			defer wg.Done()
			if err := s.Stop(); err != nil {
				this.errors <- err
			}
		}(s)
	}
}

func (this *app) signalHandler(wg *sync.WaitGroup) {
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			// this ensures a subsequent INT/TERM will trigger standard go behaviour of
			// terminating.
			signal.Stop(ch)
			this.term(wg)
			return
		}
	}
}

// Used for pretty printing addresses.
func pprintAddr(listeners []net.Listener) []byte {
	var out bytes.Buffer
	for i, l := range listeners {
		if i != 0 {
			fmt.Fprint(&out, ", ")
		}
		fmt.Fprint(&out, l.Addr())
	}
	return out.Bytes()
}
