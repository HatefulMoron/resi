package internal

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"testing"

	protos "github.com/hatefulmoron/resi/protos"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func serveEndServer(addr string) (*grpc.Server, error) {
	tcp, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer()
	protos.RegisterEsiServer(s, &EndServer{})

	go func() {
		if err = s.Serve(tcp); err != nil {
			return
		}
	}()
	return s, nil
}

func vanillaTcp(l *ResiListener, t *testing.T) {
	expected := make([]byte, 48)
	rand.Read(expected)

	assert.Equal(t, "127.0.0.1:7000", l.TcpAddr().String())

	client, err := backClient("127.0.0.1:7000")
	assert.Equal(t, nil, err)

	resp, err := client.Test(context.Background(), &protos.TestRequest{
		Ext:  nil,
		Data: expected,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, expected, resp.Data)
}

func vanillaNKN(l *ResiListener, t *testing.T) {
	expected := make([]byte, 48)
	rand.Read(expected)

	client, err := backClient(l.NKNAddr().String())
	assert.Equal(t, nil, err)

	resp, err := client.Test(context.Background(), &protos.TestRequest{
		Ext:  nil,
		Data: expected,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, expected, resp.Data)
}

func TestListenerVanillaTCP(t *testing.T) {
	s, err := serveEndServer("localhost:7099")
	assert.Equal(t, nil, err)
	defer s.Stop()

	l, err := NewResiListener(ResiListenerOptions{
		Tcp: SocketAddr{
			Net: "tcp",
			Str: "127.0.0.1:7000",
		},
		TcpForward: SocketAddr{
			Net: "tcp",
			Str: "127.0.0.1:7099",
		},
	})
	assert.Equal(t, nil, err)

	go l.Serve()
	vanillaTcp(l, t)

	l.Stop()
}

func TestListenerVanillaNKN(t *testing.T) {
	s, err := serveEndServer("localhost:7099")
	assert.Equal(t, nil, err)
	defer s.Stop()

	l, err := NewResiListener(ResiListenerOptions{
		Tcp: SocketAddr{
			Net: "tcp",
			Str: "127.0.0.1:7000",
		},
		TcpForward: SocketAddr{
			Net: "tcp",
			Str: "127.0.0.1:7099",
		},
	})
	assert.Equal(t, nil, err)

	go l.Serve()
	vanillaTcp(l, t)

	l.Stop()
}

func TestListenerVanillaBoth(t *testing.T) {
	s, err := serveEndServer("localhost:7099")
	assert.Equal(t, nil, err)
	defer s.Stop()

	l, err := NewResiListener(ResiListenerOptions{
		Tcp: SocketAddr{
			Net: "tcp",
			Str: "127.0.0.1:7000",
		},
		TcpForward: SocketAddr{
			Net: "tcp",
			Str: "127.0.0.1:7099",
		},
		NknSeed: "039e481266e5a05168c1d834a94db512dbc235877f150c5a3cc1e3903662d673",
	})
	assert.Equal(t, nil, err)

	go l.Serve()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		vanillaTcp(l, t)
		wg.Done()
	}()
	go func() {
		vanillaNKN(l, t)
		wg.Done()
	}()

	wg.Wait()

	l.Stop()
}
