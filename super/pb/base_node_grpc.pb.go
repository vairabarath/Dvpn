// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: base_node.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	BaseNodeService_RegisterSuperNode_FullMethodName   = "/dvpn.BaseNodeService/RegisterSuperNode"
	BaseNodeService_SuperNodeHeartbeat_FullMethodName  = "/dvpn.BaseNodeService/SuperNodeHeartbeat"
	BaseNodeService_GetActiveSuperNodes_FullMethodName = "/dvpn.BaseNodeService/GetActiveSuperNodes"
	BaseNodeService_RequestExitRegion_FullMethodName   = "/dvpn.BaseNodeService/RequestExitRegion"
)

// BaseNodeServiceClient is the client API for BaseNodeService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BaseNodeServiceClient interface {
	RegisterSuperNode(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	SuperNodeHeartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*Ack, error)
	GetActiveSuperNodes(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*SuperNodeList, error)
	RequestExitRegion(ctx context.Context, in *ExitRegionRequest, opts ...grpc.CallOption) (*SuperNodeList, error)
}

type baseNodeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBaseNodeServiceClient(cc grpc.ClientConnInterface) BaseNodeServiceClient {
	return &baseNodeServiceClient{cc}
}

func (c *baseNodeServiceClient) RegisterSuperNode(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, BaseNodeService_RegisterSuperNode_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *baseNodeServiceClient) SuperNodeHeartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*Ack, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Ack)
	err := c.cc.Invoke(ctx, BaseNodeService_SuperNodeHeartbeat_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *baseNodeServiceClient) GetActiveSuperNodes(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*SuperNodeList, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SuperNodeList)
	err := c.cc.Invoke(ctx, BaseNodeService_GetActiveSuperNodes_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *baseNodeServiceClient) RequestExitRegion(ctx context.Context, in *ExitRegionRequest, opts ...grpc.CallOption) (*SuperNodeList, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SuperNodeList)
	err := c.cc.Invoke(ctx, BaseNodeService_RequestExitRegion_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BaseNodeServiceServer is the server API for BaseNodeService service.
// All implementations must embed UnimplementedBaseNodeServiceServer
// for forward compatibility.
type BaseNodeServiceServer interface {
	RegisterSuperNode(context.Context, *RegisterRequest) (*RegisterResponse, error)
	SuperNodeHeartbeat(context.Context, *HeartbeatRequest) (*Ack, error)
	GetActiveSuperNodes(context.Context, *emptypb.Empty) (*SuperNodeList, error)
	RequestExitRegion(context.Context, *ExitRegionRequest) (*SuperNodeList, error)
	mustEmbedUnimplementedBaseNodeServiceServer()
}

// UnimplementedBaseNodeServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedBaseNodeServiceServer struct{}

func (UnimplementedBaseNodeServiceServer) RegisterSuperNode(context.Context, *RegisterRequest) (*RegisterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterSuperNode not implemented")
}
func (UnimplementedBaseNodeServiceServer) SuperNodeHeartbeat(context.Context, *HeartbeatRequest) (*Ack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SuperNodeHeartbeat not implemented")
}
func (UnimplementedBaseNodeServiceServer) GetActiveSuperNodes(context.Context, *emptypb.Empty) (*SuperNodeList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetActiveSuperNodes not implemented")
}
func (UnimplementedBaseNodeServiceServer) RequestExitRegion(context.Context, *ExitRegionRequest) (*SuperNodeList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestExitRegion not implemented")
}
func (UnimplementedBaseNodeServiceServer) mustEmbedUnimplementedBaseNodeServiceServer() {}
func (UnimplementedBaseNodeServiceServer) testEmbeddedByValue()                         {}

// UnsafeBaseNodeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BaseNodeServiceServer will
// result in compilation errors.
type UnsafeBaseNodeServiceServer interface {
	mustEmbedUnimplementedBaseNodeServiceServer()
}

func RegisterBaseNodeServiceServer(s grpc.ServiceRegistrar, srv BaseNodeServiceServer) {
	// If the following call pancis, it indicates UnimplementedBaseNodeServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&BaseNodeService_ServiceDesc, srv)
}

func _BaseNodeService_RegisterSuperNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BaseNodeServiceServer).RegisterSuperNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BaseNodeService_RegisterSuperNode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BaseNodeServiceServer).RegisterSuperNode(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BaseNodeService_SuperNodeHeartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BaseNodeServiceServer).SuperNodeHeartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BaseNodeService_SuperNodeHeartbeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BaseNodeServiceServer).SuperNodeHeartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BaseNodeService_GetActiveSuperNodes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BaseNodeServiceServer).GetActiveSuperNodes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BaseNodeService_GetActiveSuperNodes_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BaseNodeServiceServer).GetActiveSuperNodes(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _BaseNodeService_RequestExitRegion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExitRegionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BaseNodeServiceServer).RequestExitRegion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BaseNodeService_RequestExitRegion_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BaseNodeServiceServer).RequestExitRegion(ctx, req.(*ExitRegionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BaseNodeService_ServiceDesc is the grpc.ServiceDesc for BaseNodeService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BaseNodeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dvpn.BaseNodeService",
	HandlerType: (*BaseNodeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterSuperNode",
			Handler:    _BaseNodeService_RegisterSuperNode_Handler,
		},
		{
			MethodName: "SuperNodeHeartbeat",
			Handler:    _BaseNodeService_SuperNodeHeartbeat_Handler,
		},
		{
			MethodName: "GetActiveSuperNodes",
			Handler:    _BaseNodeService_GetActiveSuperNodes_Handler,
		},
		{
			MethodName: "RequestExitRegion",
			Handler:    _BaseNodeService_RequestExitRegion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "base_node.proto",
}
