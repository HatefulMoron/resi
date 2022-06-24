package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
	protos "github.com/hatefulmoron/resi/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RedundantClient struct {
	tcpConn   net.Conn
	tcpClient protos.EsiClient
	nknConn   net.Conn
	nknClient protos.EsiClient
	socket    net.Conn // instance of *Socket, or ptr to one of the raw sockets
	inner     protos.EsiClient
}

type RedunantAddr struct {
	tcp *string
	nkn *string
}

func newBasicClient(addr net.Addr) (net.Conn, protos.EsiClient, error) {
	connCh := make(chan net.Conn)
	var err error
	var grpcConn *grpc.ClientConn

	if addr.Network() == "tcp" {
		tcp_opt := grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				rawConn, err := net.Dial("tcp", s)
				conn := NewSocket([]net.Conn{rawConn})
				//conn.Debug = true

				if err != nil {
					return nil, err
				}

				connCh <- conn
				return conn, nil
			})

		grpcConn, err = grpc.Dial(addr.String(),
			tcp_opt,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		nkn_opt := grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				rawConn, err := Dial(s)
				conn := NewSocket([]net.Conn{rawConn})
				//conn.Debug = true

				if err != nil {
					return nil, err
				}

				connCh <- conn
				return conn, nil
			})
		grpcConn, err = grpc.Dial(addr.String(),
			nkn_opt,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if err != nil {
		return nil, nil, err
	}

	return <-connCh, protos.NewEsiClient(grpcConn), nil
}

func (c *RedundantClient) handshake() error {
	// Encode a cookie request RESI extension message
	wData, err := proto.Marshal(&protos.RedundancyCookieRequest{})
	if err != nil {
		return err
	}

	// Modify on TCP to get a cookie
	resp, err := c.tcpClient.Modify(context.Background(), &protos.ModifyRequest{
		Ext: &protos.Extension{
			Id:   0,
			Pass: false,
			Data: wData,
		},
	})
	if err != nil {
		return err
	}

	if resp.Ext == nil {
		panic("no ext")
	}

	// Decode as a cookie response
	var cookieResp protos.RedundancyResponse
	err = proto.Unmarshal(resp.Ext.Data, &cookieResp)
	if err != nil {
		return err
	}

	log.Printf("cookie: %s\n", cookieResp.Cookie)

	wData, err = proto.Marshal(&protos.RedundancyMergeRequest{
		Cookie: cookieResp.Cookie,
	})
	if err != nil {
		return err
	}
	resp, err = c.nknClient.Modify(context.Background(), &protos.ModifyRequest{
		Ext: &protos.Extension{
			Id:   1,
			Pass: false,
			Data: wData,
		},
	})
	if err != nil {
		return err
	}

	time.Sleep(time.Duration(10) * time.Second)

	(c.tcpConn.(*Socket)).Absorb(c.nknConn.(*Socket))

	return nil
}

func RedundantDial(addr RedunantAddr) (*RedundantClient, error) {
	var tcpConn net.Conn
	var nknConn net.Conn
	var tcpClient protos.EsiClient
	var nknClient protos.EsiClient
	var err error

	tcpConn = nil
	nknConn = nil
	tcpClient = nil
	nknClient = nil

	if addr.tcp != nil {
		tcpConn, tcpClient, err = newBasicClient(&SocketAddr{
			network: "tcp",
			str:     *addr.tcp,
		})

		if err != nil {
			return nil, err
		}
	}

	if addr.nkn != nil {
		nknConn, nknClient, err = newBasicClient(&SocketAddr{
			network: "nkn",
			str:     *addr.nkn,
		})

		if err != nil {
			return nil, err
		}
	}

	if tcpConn == nil && nknConn == nil {
		panic(*addr.tcp)
	}

	client := &RedundantClient{
		tcpConn:   tcpConn,
		nknConn:   nknConn,
		tcpClient: tcpClient,
		nknClient: nknClient,
	}

	if tcpConn != nil && nknConn != nil {
		err = client.handshake()
		if err != nil {
			return nil, err
		}
		client.inner = tcpClient
	} else {
		if tcpConn != nil {
			client.inner = tcpClient
		} else {
			client.inner = nknClient
		}
	}

	return client, nil
}

func (c *RedundantClient) Inner() protos.EsiClient {
	return c.inner
}
