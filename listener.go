package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	protos "github.com/hatefulmoron/resi/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

type ResiListener struct {
	server   *ResiServer
	listener net.Listener
	client   protos.EsiClient

	protos.UnimplementedEsiServer
}

func (l *ResiListener) Close() {
	l.listener.Close()
}

func (l *ResiListener) Addr() net.Addr {
	return l.listener.Addr()
}

func (l *ResiListener) Modify(
	context.Context,
	*protos.ModifyRequest,
) (*protos.ModifyResponse, error) {
	return &protos.ModifyResponse{
		Ext: nil,
	}, nil
}

func (l *ResiListener) Test(
	ctx context.Context,
	req *protos.TestRequest,
) (*protos.TestResponse, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		panic(ok)
	}

	log.Printf("%s: forwarding req %d", peer.Addr, len(req.Data))

	resp, err := l.client.Test(context.Background(), req)
	log.Printf("%s: got resp %d\n", peer, len(resp.Data))

	return resp, err
}

func backClient(addr string) (protos.EsiClient, error) {
	var protocol string
	if strings.Contains(addr, ":") {
		protocol = "tcp"
	} else {
		protocol = "nkn"
	}

	var conn *grpc.ClientConn
	var err error

	if protocol == "tcp" {
		conn, err = grpc.Dial(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		nkn_opt := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			conn, err := Dial(s)
			if err != nil {
				return nil, err
			}
			return conn, nil
		})
		conn, err = grpc.Dial(addr,
			nkn_opt,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if err != nil {
		return nil, err
	}

	return protos.NewEsiClient(conn), nil
}

func NewTCPEndpoint(server *ResiServer, back string, port int) (*ResiListener, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	client, err := backClient(back)
	if err != nil {
		return nil, err
	}

	resi := &ResiListener{
		server:   server,
		listener: l,
		client:   client,
	}

	return resi, nil
}

func NewNKNEndpoint(server *ResiServer, back string) (*ResiListener, error) {
	l, err := NewNKNListener()
	if err != nil {
		return nil, err
	}

	client, err := backClient(back)
	if err != nil {
		return nil, err
	}

	resi := &ResiListener{
		server:   server,
		listener: l,
		client:   client,
	}

	return resi, nil
}
