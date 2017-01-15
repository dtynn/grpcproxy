package client

import (
	stdctx "context"
	"log"

	"google.golang.org/grpc"

	"github.com/dtynn/grpcproxy/example/rpc"
)

func Bar(host string, say string) error {
	// creds, err := credentials.NewClientTLSFromFile("./ca.pem", "x.test.youtube.com")
	// if err != nil {
	// 	return err
	// }

	conn, err := grpc.Dial(host, grpc.WithInsecure())
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
