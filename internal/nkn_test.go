package internal

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"testing"

	helloworld "github.com/hatefulmoron/resi/protos"
	"github.com/stretchr/testify/assert"
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

func TestNKNSession(t *testing.T) {

	listener, err := NewNKNListener("039e481266e5a05168c1d834a94db512dbc235877f150c5a3cc1e3903672c683")
	assert.Equal(t, nil, err)

	addr := listener.Addr()

	go func() {
		fd, err := listener.Accept()
		assert.Equal(t, nil, err)

		buf := make([]byte, 1024)
		for i := 0; i < 100; i++ {
			n, err := fd.Read(buf)
			if err != nil {
				break
			}

			n2, err := fd.Write(buf[:n])
			assert.Equal(t, nil, err)
			assert.Equal(t, n, n2)
		}
	}()

	fd, err := Dial(addr.String())
	assert.Equal(t, nil, err)

	for i := 0; i < 50; i++ {
		log.Printf("%d\n", i)
		expected := make([]byte, 1024)
		rand.Read(expected)

		n, err := fd.Write(expected)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(expected), n)

		buf := make([]byte, len(expected))

		n, err = fd.Read(buf)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(expected), n)
		assert.Equal(t, expected, buf)
	}

}

func TestNKNGRPC(t *testing.T) {
	c := make(chan net.Addr)
	f := make(chan struct{})
	errors := make(chan string)
	s := grpc.NewServer()

	go func() {
		listener, err := NewNKNListener("039e481266e5a05168c1d834a94db512dbc235877f150c5a3cc1e3903672c683")
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
