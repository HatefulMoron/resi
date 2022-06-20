package main

import (
	"context"
	"log"
	"testing"

	protos "github.com/hatefulmoron/resi/protos"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newClient(addr string) protos.EsiClient {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	c := protos.NewEsiClient(conn)
	return c
}

func TestVanillaTCP(t *testing.T) {
	endStop := make(chan struct{})
	endL, _ := StartEndServer(9090, endStop)
	defer endL.Close()

	srv := StartResiServer(9091, "127.0.0.1:9090")
	defer func() {
		srv.Close()

		for {
			log.Println("waiting")
			l, ok := <-srv.Listener
			if !ok {
				return
			}
			l.Close()
		}
	}()

	c := newClient("localhost:9091")

	data := []byte{1, 2, 3, 4, 5}

	resp, err := c.Test(context.Background(), &protos.TestRequest{
		Data: data,
	})

	assert.Equal(t, nil, err)
	assert.Equal(t, resp.Data, data)
}

func TestVanillaInterfaces(t *testing.T) {
	endStop := make(chan struct{})
	endL, _ := StartEndServer(9092, endStop)
	defer endL.Close()

	srv := StartResiServer(9093, "127.0.0.1:9092")

	ls := make([]*ResiListener, 0)
	for {
		if len(ls) == 2 {
			break
		}

		l, ok := <-srv.Listener
		if !ok {
			break
		}

		ls = append(ls, l)
	}

	for _, v := range ls {
		v.Close()
	}
}

func TestVanillaNKN(t *testing.T) {
	endStop := make(chan struct{})
	endL, _ := StartEndServer(9098, endStop)
	defer endL.Close()

	srv := StartResiServer(9099, "127.0.0.1:9098")

	ls := make([]*ResiListener, 0)
	for {
		if len(ls) == 2 {
			break
		}

		l, ok := <-srv.Listener
		if !ok {
			break
		}

		ls = append(ls, l)
	}

	for _, v := range ls {
		v.Close()
	}
}
