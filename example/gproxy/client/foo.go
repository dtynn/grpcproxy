package client

import (
	stdctx "context"
	"log"

	"google.golang.org/grpc"
	// "google.golang.org/grpc/credentials"

	"github.com/dtynn/grpcproxy/example/rpc"
)

func Foo(host string, hello string) error {
	// creds, err := credentials.NewClientTLSFromFile("./ca.pem", "x.test.youtube.com")
	// if err != nil {
	// 	return err
	// }

	conn, err := grpc.Dial(host, grpc.WithInsecure())
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
