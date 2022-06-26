package internal

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
	Tcp *string
	Nkn *string
}

func newBasicClient(addr net.Addr) (net.Conn, protos.EsiClient, error) {
	connCh := make(chan net.Conn)
	var err error
	var grpcConn *grpc.ClientConn

	if addr.Network() == "tcp" {
		tcp_opt := grpc.WithContextDialer(
			func(ctx context.Context, s string) (net.Conn, error) {
				rawConn, err := net.Dial("tcp", s)
				if err != nil {
					return nil, err
				}

				conn := NewSocket([]net.Conn{rawConn})
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
				if err != nil {
					return nil, err
				}

				conn := NewSocket([]net.Conn{rawConn})
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
	log.Println("asking for merging")

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

	time.Sleep(5 * time.Second)
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

	if addr.Tcp != nil {
		log.Println("conn tcp")
		tcpConn, tcpClient, err = newBasicClient(&SocketAddr{
			Net: "tcp",
			Str: *addr.Tcp,
		})

		if err != nil {
			return nil, err
		}
	}

	if addr.Nkn != nil {
		log.Println("conn nkn")
		nknConn, nknClient, err = newBasicClient(&SocketAddr{
			Net: "nkn",
			Str: *addr.Nkn,
		})

		if err != nil {
			return nil, err
		}
	}

	if tcpConn == nil && nknConn == nil {
		panic(*addr.Tcp)
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
		log.Println("inner tcp (combined)")
	} else {
		if tcpConn != nil {
			client.inner = tcpClient
			log.Println("inner tcp")
		} else {
			client.inner = nknClient
			log.Println("inner nkn")
		}
	}

	return client, nil
}

func (c *RedundantClient) Inner() protos.EsiClient {
	return c.inner
}

func (c *RedundantClient) Addrs() []net.Addr {
	ret := make([]net.Addr, 0)

	if c.tcpConn != nil {
		ret = append(ret, c.tcpConn.(*Socket).LocalAddrs()...)
	}
	if c.nknConn != nil {
		ret = append(ret, c.nknConn.(*Socket).LocalAddrs()...)
	}

	return ret
}
