package client

import (
	stdctx "context"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/dtynn/grpcproxy/example/rpc"
)

func Pulse(host string, greeting string, cert, hostname string) error {
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

	cli := rpc.NewPulseClient(conn)

	times := 10

	pieces := strings.Split(greeting, ";")

	if len(pieces) == 2 {
		greeting = pieces[0]
		t, _ := strconv.Atoi(pieces[1])
		if t > 0 {
			times = t
		}
	}

	stream, err := cli.Beat(stdctx.Background())
	if err != nil {
		return err
	}

	waitc := make(chan error, times*2+1)

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				waitc <- err
				return
			}

			log.Printf("got beat resp: %s", in.Received)
		}
	}()

	for i := 0; i < times; i++ {
		if err := stream.Send(&rpc.BeatReq{
			Greeting: fmt.Sprintf("#%d %s", i+1, greeting),
		}); err != nil {
			return err
		}
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}

	return <-waitc
}
