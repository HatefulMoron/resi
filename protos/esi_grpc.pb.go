// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package protos

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// EsiClient is the client API for Esi service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EsiClient interface {
	Modify(ctx context.Context, in *ModifyRequest, opts ...grpc.CallOption) (*ModifyResponse, error)
	Test(ctx context.Context, in *TestRequest, opts ...grpc.CallOption) (*TestResponse, error)
}

type esiClient struct {
	cc grpc.ClientConnInterface
}

func NewEsiClient(cc grpc.ClientConnInterface) EsiClient {
	return &esiClient{cc}
}

func (c *esiClient) Modify(ctx context.Context, in *ModifyRequest, opts ...grpc.CallOption) (*ModifyResponse, error) {
	out := new(ModifyResponse)
	err := c.cc.Invoke(ctx, "/protos.Esi/Modify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *esiClient) Test(ctx context.Context, in *TestRequest, opts ...grpc.CallOption) (*TestResponse, error) {
	out := new(TestResponse)
	err := c.cc.Invoke(ctx, "/protos.Esi/Test", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EsiServer is the server API for Esi service.
// All implementations must embed UnimplementedEsiServer
// for forward compatibility
type EsiServer interface {
	Modify(context.Context, *ModifyRequest) (*ModifyResponse, error)
	Test(context.Context, *TestRequest) (*TestResponse, error)
	mustEmbedUnimplementedEsiServer()
}

// UnimplementedEsiServer must be embedded to have forward compatible implementations.
type UnimplementedEsiServer struct {
}

func (UnimplementedEsiServer) Modify(context.Context, *ModifyRequest) (*ModifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Modify not implemented")
}
func (UnimplementedEsiServer) Test(context.Context, *TestRequest) (*TestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Test not implemented")
}
func (UnimplementedEsiServer) mustEmbedUnimplementedEsiServer() {}

// UnsafeEsiServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EsiServer will
// result in compilation errors.
type UnsafeEsiServer interface {
	mustEmbedUnimplementedEsiServer()
}

func RegisterEsiServer(s grpc.ServiceRegistrar, srv EsiServer) {
	s.RegisterService(&Esi_ServiceDesc, srv)
}

func _Esi_Modify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EsiServer).Modify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.Esi/Modify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EsiServer).Modify(ctx, req.(*ModifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Esi_Test_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EsiServer).Test(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.Esi/Test",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EsiServer).Test(ctx, req.(*TestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Esi_ServiceDesc is the grpc.ServiceDesc for Esi service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Esi_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protos.Esi",
	HandlerType: (*EsiServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Modify",
			Handler:    _Esi_Modify_Handler,
		},
		{
			MethodName: "Test",
			Handler:    _Esi_Test_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/esi.proto",
}
