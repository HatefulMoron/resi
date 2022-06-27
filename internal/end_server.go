package internal

import (
	"context"
	"log"

	protos "github.com/hatefulmoron/resi/protos"
)

type EndServer struct {
	protos.UnimplementedEsiServer
}

func (s *EndServer) Test(
	ctx context.Context,
	req *protos.TestRequest,
) (*protos.TestResponse, error) {
	log.Printf("responding to request: %v\n", req.Data)
	return &protos.TestResponse{
		Ext:  nil,
		Data: req.Data,
	}, nil
}
