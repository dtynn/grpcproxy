package client

import (
	stdctx "context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/dtynn/grpcproxy/example/rpc"
)

func Foo(host string, hello string, cert, hostname string) error {
	var opts []grpc.DialOption
	if cert != "" && hostname != "" {
		creds, err := credentials.NewClientTLSFromFile(cert, hostname)
		if err != nil {
			return err
		}

		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(host, opts...)
	if err != nil {
		return err
	}

	defer conn.Close()

	cli := rpc.NewFooClient(conn)

	resp, err := cli.Chat(stdctx.Background(), &rpc.FooReq{Hello: hello})
	if err != nil {
		return err
	}

	log.Printf("got foo response %q", resp.World)
	return nil
}
