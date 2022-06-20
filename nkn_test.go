package main

import (
	"context"
	"fmt"
	"net"
	"testing"

	helloworld "github.com/hatefulmoron/resi/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	helloworld.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{
		Message: "hey i got that",
	}, nil
}

func TestNKNGRPC(t *testing.T) {
	c := make(chan net.Addr)
	f := make(chan struct{})
	errors := make(chan string)
	s := grpc.NewServer()

	go func() {
		listener, err := NewNKNListener()
		if err != nil {
			errors <- fmt.Sprintf("new listener: %v", err)
			return
		}

		c <- listener.Addr()

		helloworld.RegisterGreeterServer(s, &server{})
		if err = s.Serve(listener); err != nil {
			errors <- fmt.Sprintf("failed to serve: %v", err)
			return
		}
	}()

	go func() {
		addr := <-c

		opt := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			conn, err := Dial(s)
			if err != nil {
				return nil, err
			}
			return conn, nil
		})

		conn, err := grpc.Dial(addr.String(),
			opt,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			errors <- fmt.Sprintf("failed to dial: %v", err)
			return
		}

		c := helloworld.NewGreeterClient(conn)
		_, err = c.SayHello(context.Background(), &helloworld.HelloRequest{
			Name: "Thomas",
		})

		if err != nil {
			errors <- fmt.Sprintf("rpc fail: %v", err)
			return
		}

		// ..
		f <- struct{}{}
		conn.Close()
	}()

	select {
	case <-f:
		s.Stop()
		return
	case e := <-errors:
		t.Fatal(e)
	}
}
