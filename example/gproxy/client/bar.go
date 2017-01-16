package client

import (
	stdctx "context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/dtynn/grpcproxy/example/rpc"
)

func Bar(host string, say string, cert, hostname string) error {
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

	cli := rpc.NewBarClient(conn)

	resp, err := cli.Talk(stdctx.Background(), &rpc.BarReq{Say: say})
	if err != nil {
		return err
	}

	log.Printf("got bar response %q", resp.Hear)
	return nil
}
