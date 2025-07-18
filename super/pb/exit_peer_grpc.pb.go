// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: exit_peer.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ExitPeerService_GetWireGuardInfo_FullMethodName = "/dvpn.ExitPeerService/GetWireGuardInfo"
)

// ExitPeerServiceClient is the client API for ExitPeerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExitPeerServiceClient interface {
	GetWireGuardInfo(ctx context.Context, in *ExitPeerInfoRequest, opts ...grpc.CallOption) (*ExitPeerInfoResponse, error)
}

type exitPeerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewExitPeerServiceClient(cc grpc.ClientConnInterface) ExitPeerServiceClient {
	return &exitPeerServiceClient{cc}
}

func (c *exitPeerServiceClient) GetWireGuardInfo(ctx context.Context, in *ExitPeerInfoRequest, opts ...grpc.CallOption) (*ExitPeerInfoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ExitPeerInfoResponse)
	err := c.cc.Invoke(ctx, ExitPeerService_GetWireGuardInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExitPeerServiceServer is the server API for ExitPeerService service.
// All implementations must embed UnimplementedExitPeerServiceServer
// for forward compatibility.
type ExitPeerServiceServer interface {
	GetWireGuardInfo(context.Context, *ExitPeerInfoRequest) (*ExitPeerInfoResponse, error)
	mustEmbedUnimplementedExitPeerServiceServer()
}

// UnimplementedExitPeerServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedExitPeerServiceServer struct{}

func (UnimplementedExitPeerServiceServer) GetWireGuardInfo(context.Context, *ExitPeerInfoRequest) (*ExitPeerInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWireGuardInfo not implemented")
}
func (UnimplementedExitPeerServiceServer) mustEmbedUnimplementedExitPeerServiceServer() {}
func (UnimplementedExitPeerServiceServer) testEmbeddedByValue()                         {}

// UnsafeExitPeerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExitPeerServiceServer will
// result in compilation errors.
type UnsafeExitPeerServiceServer interface {
	mustEmbedUnimplementedExitPeerServiceServer()
}

func RegisterExitPeerServiceServer(s grpc.ServiceRegistrar, srv ExitPeerServiceServer) {
	// If the following call pancis, it indicates UnimplementedExitPeerServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ExitPeerService_ServiceDesc, srv)
}

func _ExitPeerService_GetWireGuardInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExitPeerInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExitPeerServiceServer).GetWireGuardInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ExitPeerService_GetWireGuardInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExitPeerServiceServer).GetWireGuardInfo(ctx, req.(*ExitPeerInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ExitPeerService_ServiceDesc is the grpc.ServiceDesc for ExitPeerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ExitPeerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dvpn.ExitPeerService",
	HandlerType: (*ExitPeerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetWireGuardInfo",
			Handler:    _ExitPeerService_GetWireGuardInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "exit_peer.proto",
}
