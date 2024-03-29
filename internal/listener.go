package internal

import (
	"context"
	"errors"
	"log"
	"net"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	protos "github.com/hatefulmoron/resi/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

type ResiListenerOptions struct {
	Tcp        net.Addr
	NknSeed    string
	TcpForward net.Addr
	NknForward net.Addr
}

type ResiListener struct {
	socketListener *SocketListener
	socketCh       chan net.Conn
	sockets        map[string]*Socket // id -> socket
	peerMap        map[string]string  // localAddr -> id
	tcpAddr        net.Addr
	nknAddr        net.Addr
	server         *grpc.Server
	forward        *RedundantClient

	protos.UnimplementedEsiServer
}

func (l *ResiListener) Serve() {
	protos.RegisterEsiServer(l.server, l)

	if err := l.server.Serve(l.socketListener); err != nil {
		return // TODO?
	}
}

func (l *ResiListener) Stop() {
	l.socketListener.Close()
	l.server.Stop()
}

func (l *ResiListener) TcpAddr() net.Addr {
	return l.tcpAddr
}

func (l *ResiListener) NKNAddr() net.Addr {
	return l.nknAddr
}

func (l *ResiListener) Connected(fd *Socket) {
	id := randId(32)
	l.sockets[id] = fd
	l.peerMap[fd.RemoteAddr().String()] = id
}

func (l *ResiListener) GetSocket(ctx context.Context) (*Socket, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("todo")
	}

	id := l.peerMap[peer.Addr.String()]
	return l.sockets[id], nil
}

func (l *ResiListener) Modify(
	ctx context.Context,
	req *protos.ModifyRequest,
) (*protos.ModifyResponse, error) {

	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("peer failed")
	}

	// RESI RedundancyCookieRequest
	if req.Ext != nil && req.Ext.Id == 0 {

		wData, err := proto.Marshal(&protos.RedundancyResponse{
			Addr:   "",
			Cookie: l.peerMap[peer.Addr.String()],
		})
		if err != nil {
			return nil, err
		}

		return &protos.ModifyResponse{
			Ext: &protos.Extension{
				Id:   1,
				Pass: false,
				Data: wData,
			},
		}, nil
	} else if req.Ext.Id == 1 {

		// RESI RedundancyMergeRequest

		var resiReq protos.RedundancyMergeRequest
		err := proto.Unmarshal(req.Ext.Data, &resiReq)
		if err != nil {
			return nil, err
		}

		masterCookie := resiReq.Cookie
		slaveCookie := l.peerMap[peer.Addr.String()]

		if masterCookie == slaveCookie {
			panic("todo") // you shouldn't be sending this on this transport..
		}

		masterSocket := l.sockets[masterCookie]
		slaveSocket := l.sockets[slaveCookie]

		log.Printf("absorbing %s into %s\n", slaveCookie, masterCookie)

		//masterSocket.AbsorbCb = func(master *Socket) error {
		//	log.Println("slkdjlsjfd")
		//	master.RawAbsorb(slaveSocket)
		//	return nil
		//}

		go func() {
			time.Sleep(5 * time.Second)
			masterSocket.Absorb(slaveSocket)
		}()

	}

	return &protos.ModifyResponse{
		Ext: nil,
	}, nil
}

func (l *ResiListener) Test(
	ctx context.Context,
	req *protos.TestRequest,
) (*protos.TestResponse, error) {
	//peer, ok := peer.FromContext(ctx)
	_, ok := peer.FromContext(ctx)
	if !ok {
		panic(ok)
	}

	//log.Printf("%s: forwarding req %d", peer.Addr, len(req.Data))
	return l.forward.Inner().Test(ctx, req)
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
		tcp_opt := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			conn, err := net.Dial("tcp", s)
			if err != nil {
				return nil, err
			}
			return NewSocket([]net.Conn{conn}), nil
		})
		conn, err = grpc.Dial(addr, tcp_opt,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		nkn_opt := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			conn, err := Dial(s)
			if err != nil {
				return nil, err
			}
			return NewSocket([]net.Conn{conn}), nil
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

func makeTcpListener(tcpAddr net.Addr) (net.Listener, error) {
	l, err := net.Listen("tcp", tcpAddr.String())
	return l, err
}

func makeNKNListener(seed string) (net.Listener, error) {
	l, err := NewNKNListener(seed)
	return l, err
}

func NewResiListener(options ResiListenerOptions) (*ResiListener, error) {
	listener := &ResiListener{
		socketCh: make(chan net.Conn, 16),
		sockets:  make(map[string]*Socket),
		peerMap:  make(map[string]string),
		server:   grpc.NewServer(),
	}

	forward, err := RedundantDial(RedunantAddr{
		Tcp: options.TcpForward,
		Nkn: options.NknForward,
	})

	if err != nil {
		log.Fatalf("failed to connect: %v\n", err)
		return nil, err
	}

	listener.forward = forward

	listeners := make([]net.Listener, 0)

	if options.Tcp != nil {

		tcp, err := makeTcpListener(options.Tcp)
		if err != nil {
			return nil, err
		}

		listeners = append(listeners, tcp)
		listener.tcpAddr = tcp.Addr()

	}

	if options.NknSeed != "" {

		nkn, err := makeNKNListener(options.NknSeed)
		if err != nil {
			return nil, err
		}

		listeners = append(listeners, nkn)
		listener.nknAddr = nkn.Addr()

	}

	listener.socketListener = NewSocketListener(listeners)
	listener.socketListener.SetAcceptObserver(listener.socketCh)

	go func() {
		for {
			fd, ok := <-listener.socketCh
			if !ok {
				break
			}

			listener.Connected(fd.(*Socket))
		}
	}()

	return listener, nil
}
