package internal

import (
	"context"

	protos "github.com/hatefulmoron/resi/protos"
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
