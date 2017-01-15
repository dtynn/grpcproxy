package bar

import (
	"fmt"
	"log"
	"net"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/dtynn/grpcproxy/example/rpc"
)

var _ rpc.BarServer = &Service{}

type Service struct {
	Port int
}

func (this *Service) Talk(ctx context.Context, req *rpc.BarReq) (*rpc.BarResp, error) {
	log.Printf("talk requested %s", req.Say)

	if strings.Contains(req.Say, "error") {
		return nil, fmt.Errorf("[BAR ON %d]bar.Talk: got talk error %s", this.Port, req.Say)
	}

	return &rpc.BarResp{
		Hear: fmt.Sprintf("[BAR ON %d]bar.Talk: Hear saying %s", this.Port, req.Say),
	}, nil
}

func (this *Service) Run() error {
	addr := fmt.Sprintf(":%d", this.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	rpc.RegisterBarServer(server, this)

	log.Printf("listen on %s", addr)

	return server.Serve(lis)
}
