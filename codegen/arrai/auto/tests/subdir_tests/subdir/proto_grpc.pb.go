// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package subdir

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

// GRPC_SubdirClient is the client API for GRPC_Subdir service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GRPC_SubdirClient interface {
	Sub(ctx context.Context, in *SubdirRequest, opts ...grpc.CallOption) (*SubdirReply, error)
}

type gRPC_SubdirClient struct {
	cc grpc.ClientConnInterface
}

func NewGRPC_SubdirClient(cc grpc.ClientConnInterface) GRPC_SubdirClient {
	return &gRPC_SubdirClient{cc}
}

func (c *gRPC_SubdirClient) Sub(ctx context.Context, in *SubdirRequest, opts ...grpc.CallOption) (*SubdirReply, error) {
	out := new(SubdirReply)
	err := c.cc.Invoke(ctx, "/subdir.GRPC_Subdir/Sub", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GRPC_SubdirServer is the server API for GRPC_Subdir service.
// All implementations must embed UnimplementedGRPC_SubdirServer
// for forward compatibility
type GRPC_SubdirServer interface {
	Sub(context.Context, *SubdirRequest) (*SubdirReply, error)
	mustEmbedUnimplementedGRPC_SubdirServer()
}

// UnimplementedGRPC_SubdirServer must be embedded to have forward compatible implementations.
type UnimplementedGRPC_SubdirServer struct {
}

func (UnimplementedGRPC_SubdirServer) Sub(context.Context, *SubdirRequest) (*SubdirReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sub not implemented")
}
func (UnimplementedGRPC_SubdirServer) mustEmbedUnimplementedGRPC_SubdirServer() {}

// UnsafeGRPC_SubdirServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GRPC_SubdirServer will
// result in compilation errors.
type UnsafeGRPC_SubdirServer interface {
	mustEmbedUnimplementedGRPC_SubdirServer()
}

func RegisterGRPC_SubdirServer(s grpc.ServiceRegistrar, srv GRPC_SubdirServer) {
	s.RegisterService(&GRPC_Subdir_ServiceDesc, srv)
}

func _GRPC_Subdir_Sub_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubdirRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GRPC_SubdirServer).Sub(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/subdir.GRPC_Subdir/Sub",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GRPC_SubdirServer).Sub(ctx, req.(*SubdirRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// GRPC_Subdir_ServiceDesc is the grpc.ServiceDesc for GRPC_Subdir service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var GRPC_Subdir_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "subdir.GRPC_Subdir",
	HandlerType: (*GRPC_SubdirServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Sub",
			Handler:    _GRPC_Subdir_Sub_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto.proto",
}
