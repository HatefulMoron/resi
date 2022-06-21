package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	protos "github.com/hatefulmoron/resi/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type EndServer struct {
	protos.UnimplementedEsiServer
}

func (s *EndServer) Test(
	ctx context.Context,
	req *protos.TestRequest,
) (*protos.TestResponse, error) {
	return &protos.TestResponse{
		Ext:  nil,
		Data: req.Data,
	}, nil
}

func endServer(port int) {
	log.Println("starting tcp endpoint")

	stop := make(chan struct{})
	_, err := StartEndServer(port, stop)
	if err != nil {
		panic(err)
	}

	<-stop
}

func StartEndServer(port int, stop chan struct{}) (net.Listener, error) {
	tcp, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}

	go func() {
		defer func() {
			stop <- struct{}{}
		}()
		s := grpc.NewServer()
		protos.RegisterEsiServer(s, &EndServer{})
		if err = s.Serve(tcp); err != nil {
			return
		}
	}()

	return tcp, nil
}

func client(addr string) {

	nkn_opt := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		conn, err := Dial(s)
		if err != nil {
			return nil, err
		}
		return conn, nil
	})

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
		conn, err = grpc.Dial(addr,
			nkn_opt,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if err != nil {
		log.Printf(" failed to dial: %v", err)
		return
	}

	c := protos.NewEsiClient(conn)
	resp, err := c.Test(context.Background(), &protos.TestRequest{
		Data: []byte{1, 2, 3, 4, 5},
	})

	if err != nil {
		log.Printf("failed to request: %v\n", err)
	}

	log.Printf("got data back: %v", resp.Data)
}

func StartResiServer(tcpport int, back string) *ResiServer {

	if back == "" {
		log.Println("missing b flag")
		panic(back)
	}

	log.Printf("starting proxy backing to %s\n", back)

	var protocol string
	if strings.Contains(back, ":") {
		protocol = "tcp"
	} else {
		protocol = "nkn"
	}

	log.Println(protocol)

	srv, err := NewResiServer(protocol, back)
	if err != nil {
		log.Printf("failed to connect to backing server: %v\n", err)
		panic(err)
	}

	go func() {
		s := grpc.NewServer()

		log.Println("starting tcp endpoint")
		tcp, err := NewTCPEndpoint(srv, back, tcpport)
		if err != nil {
			log.Printf("failed to start tcp endpoint: %v\n", err)
			srv.wg.Done()
			return
		}
		log.Printf("tcp addr: %s\n", tcp.Addr().String())

		srv.Listener <- tcp
		srv.wg.Done()

		protos.RegisterEsiServer(s, tcp)

		if err = s.Serve(tcp.listener); err != nil {
			log.Printf("tcp: failed to serve: %v\n", err)
			return
		}
	}()

	go func() {
		defer srv.wg.Done()
		s := grpc.NewServer()

		log.Println("starting nkn endpoint")
		nkn, err := NewNKNEndpoint(srv, back)
		if err != nil {
			log.Printf("failed to start nkn endpoint: %v\n", err)
			srv.wg.Done()
			return
		}

		log.Printf("nkn addr: %s\n", nkn.Addr().String())

		srv.Listener <- nkn
		srv.wg.Done()

		protos.RegisterEsiServer(s, nkn)

		if err = s.Serve(nkn.listener); err != nil {
			log.Printf("nkn: failed to serve: %v\n", err)
			return
		}
	}()

	return srv
}

func resiServer(port int, back string) {
	srv := StartResiServer(port, back)
	srv.Wait()

	ch := make(chan struct{})
	<-ch
}

func main() {
	app := &cli.App{
		Name:  "resi",
		Usage: "Redundant ESI Server Extension",

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "mode",
				Value: "gateway",
				Usage: "execution mode (end, gateway, client)",
			},
			&cli.StringFlag{
				Name:  "back",
				Value: "",
				Usage: "backing server TCP address",
			},
			&cli.IntFlag{
				Name:  "tcp",
				Value: 9000,
				Usage: "TCP listen port",
			},
		},

		Action: func(c *cli.Context) error {

			if c.String("mode") == "gateway" {
				resiServer(c.Int("tcp"), c.String("back"))
			} else if c.String("mode") == "end" {
				endServer(c.Int("tcp"))
			} else {
				client(c.String("back"))
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		return
	}
}
