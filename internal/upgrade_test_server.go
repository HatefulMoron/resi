package internal

import (
	"context"

	protos "github.com/hatefulmoron/resi/protos"
)

// When upgradeTestServer has sent two responses, merge the two
type upgradeTestServer struct {
	sockets []*Socket
	protos.UnimplementedGreeterServer
}

func (s *upgradeTestServer) SayHello(
	ctx context.Context,
	in *protos.HelloRequest,
) (*protos.HelloReply, error) {
	return &protos.HelloReply{
		Message: "hello world",
	}, nil
}

func (s *upgradeTestServer) Absorb() {
	s.sockets[0].Absorb(s.sockets[1])
}
