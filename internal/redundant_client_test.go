package internal

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	protos "github.com/hatefulmoron/resi/protos"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestRedundantClientVanillaNKN(t *testing.T) {
	c := make(chan net.Addr)
	s := grpc.NewServer()

	go func() {

		listener, err := NewNKNListener("039e481266e5a05168c1d834a94db512dbc235877f150c5a3cc1e3903672c683")
		if err != nil {
			panic(err)
		}

		c <- listener.Addr()

		protos.RegisterEsiServer(s, &EndServer{})
		if err = s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	client, err := RedundantDial(RedunantAddr{
		Nkn: <-c,
	})
	assert.Equal(t, nil, err)

	data := []byte{1, 2, 3, 4, 5}
	resp, err := client.Inner().Test(context.Background(), &protos.TestRequest{
		Ext:  nil,
		Data: data,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, true, bytes.Equal(data, resp.Data))

	s.Stop()
}

func TestRedundantClientVanillaTCP(t *testing.T) {
	c := make(chan net.Addr)
	s := grpc.NewServer()

	go func() {

		listener, err := net.Listen("tcp", "localhost:9050")
		if err != nil {
			panic(err)
		}

		c <- listener.Addr()

		protos.RegisterEsiServer(s, &EndServer{})
		if err = s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	client, err := RedundantDial(RedunantAddr{
		Tcp: <-c,
	})
	assert.Equal(t, nil, err)

	data := []byte{1, 2, 3, 4, 5}
	resp, err := client.Inner().Test(context.Background(), &protos.TestRequest{
		Ext:  nil,
		Data: data,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, true, bytes.Equal(data, resp.Data))

	s.Stop()
}

func TestRedundantClientBoth(t *testing.T) {
	c := make(chan net.Addr)
	s := grpc.NewServer()

	go func() {

		listener, err := net.Listen("tcp", "localhost:9050")
		if err != nil {
			panic(err)
		}

		c <- listener.Addr()

		protos.RegisterEsiServer(s, &EndServer{})
		if err = s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	l, err := NewResiListener(ResiListenerOptions{
		TcpForward: <-c,
		Tcp: SocketAddr{
			Net: "tcp",
			Str: "127.0.0.1:7000",
		},
		NknSeed: "039e481266e5a05168c1d834a94db512dbc235877f150c5a3cc1e3903672c673",
	})
	assert.Equal(t, nil, err)

	go l.Serve()

	time.Sleep(time.Second)

	client, err := RedundantDial(RedunantAddr{
		Tcp: l.TcpAddr(),
		Nkn: l.NKNAddr(),
	})
	assert.Equal(t, nil, err)

	data := []byte{1, 2, 3, 4, 5}

	resp, err := client.Inner().Test(context.Background(), &protos.TestRequest{
		Ext:  nil,
		Data: data,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, true, bytes.Equal(data, resp.Data))

	l.Stop()
	s.Stop()
}
