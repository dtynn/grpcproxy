package foo

import (
	"fmt"
	"log"
	"net"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dtynn/grpcproxy/example/rpc"
)

var _ rpc.FooServer = &Service{}

type Service struct {
	Port int
}

func (this *Service) Chat(ctx context.Context, req *rpc.FooReq) (*rpc.FooResp, error) {
	log.Printf("chat requested: %s", req.Hello)

	if strings.Contains(req.Hello, "error") {
		return nil, fmt.Errorf("[FOO ON %d]foo.Chat: got chat error %s", this.Port, req.Hello)
	}

	return &rpc.FooResp{
		World: fmt.Sprintf("[FOO ON %d]foo.Chat: Hello %s World", this.Port, req.Hello),
	}, nil
}

func (this *Service) Run() error {
	addr := fmt.Sprintf(":%d", this.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	rpc.RegisterFooServer(server, this)

	log.Printf("listen on %s", addr)

	return server.Serve(lis)
}
