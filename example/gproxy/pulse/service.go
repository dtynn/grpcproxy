package pulse

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/dtynn/grpcproxy/example/rpc"
)

var _ rpc.PulseServer = &Service{}

type Service struct {
	Port int
}

func (this *Service) Beat(stream rpc.Pulse_BeatServer) error {
	var count int
	defer func() {
		log.Printf("got %d greetings", count)
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		resp := &rpc.BeatResp{
			Received: fmt.Sprintf("[PULSE ON %d] received %q at %s", this.Port, req.Greeting, time.Now()),
		}

		if err := stream.Send(resp); err != nil {
			return err
		}

		count += 1
	}

	return nil
}

func (this *Service) Run() error {
	addr := fmt.Sprintf(":%d", this.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	rpc.RegisterPulseServer(server, this)

	log.Printf("listen on %s", addr)

	return server.Serve(lis)
}
