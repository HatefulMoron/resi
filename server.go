package main

import (
	"context"
	"net"
	"sync"

	protos "github.com/hatefulmoron/resi/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ResiServer struct {
	client       protos.EsiClient
	wg           sync.WaitGroup
	Listener     chan *ResiListener
	associations map[string]*Association // peer addr -> assoc
	lock         sync.Mutex
}

func (s *ResiServer) Close() {
	s.wg.Wait()
	close(s.Listener)
}

func (s *ResiServer) Wait() {
	s.wg.Wait()
}

func NewResiServer(protocol string, b string) (*ResiServer, error) {
	if protocol == "tcp" {
		conn, err := grpc.Dial(b,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}

		c := protos.NewEsiClient(conn)

		srv := &ResiServer{
			client:       c,
			associations: make(map[string]*Association),
			Listener:     make(chan *ResiListener, 16),
		}
		srv.wg.Add(2)

		return srv, nil

	} else if protocol == "nkn" {
		opt := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			conn, err := Dial(s)
			if err != nil {
				return nil, err
			}
			return conn, nil
		})

		conn, err := grpc.Dial(b,
			opt,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}

		c := protos.NewEsiClient(conn)
		srv := &ResiServer{
			client:       c,
			associations: make(map[string]*Association),
			Listener:     make(chan *ResiListener, 16),
		}
		srv.wg.Add(2)

		return srv, nil
	}
	panic(protocol)
}
