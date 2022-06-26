package internal

import (
	"bytes"
	"context"
	"log"
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

		listener, err := NewNKNListener()
		if err != nil {
			panic(err)
		}

		c <- listener.Addr()

		protos.RegisterEsiServer(s, &EndServer{})
		if err = s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	nkn_addr := (<-c).String()

	client, err := RedundantDial(RedunantAddr{
		Tcp: nil,
		Nkn: &nkn_addr,
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

	addr := (<-c).String()

	client, err := RedundantDial(RedunantAddr{
		Tcp: &addr,
		Nkn: nil,
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
	l, err := NewResiListener(SocketAddr{
		Net: "tcp",
		Str: "127.0.0.1:7000",
	})
	assert.Equal(t, nil, err)

	go l.Serve()

	time.Sleep(time.Second)

	tcpAddr := l.TcpAddr().String()
	nknAddr := l.NKNAddr().String()

	client, err := RedundantDial(RedunantAddr{
		Tcp: &tcpAddr,
		Nkn: &nknAddr,
	})
	assert.Equal(t, nil, err)

	data := []byte{1, 2, 3, 4, 5}

	log.Println("==================================")

	resp, err := client.Inner().Test(context.Background(), &protos.TestRequest{
		Ext:  nil,
		Data: data,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, true, bytes.Equal(data, resp.Data))

	l.Stop()
}
