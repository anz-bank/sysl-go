// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package encoder_backend

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

// EncoderBackendClient is the client API for EncoderBackend service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EncoderBackendClient interface {
	Rot13(ctx context.Context, in *EncodingRequest, opts ...grpc.CallOption) (*EncodingResponse, error)
}

type encoderBackendClient struct {
	cc grpc.ClientConnInterface
}

func NewEncoderBackendClient(cc grpc.ClientConnInterface) EncoderBackendClient {
	return &encoderBackendClient{cc}
}

func (c *encoderBackendClient) Rot13(ctx context.Context, in *EncodingRequest, opts ...grpc.CallOption) (*EncodingResponse, error) {
	out := new(EncodingResponse)
	err := c.cc.Invoke(ctx, "/encoder_backend.EncoderBackend/Rot13", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EncoderBackendServer is the server API for EncoderBackend service.
// All implementations must embed UnimplementedEncoderBackendServer
// for forward compatibility
type EncoderBackendServer interface {
	Rot13(context.Context, *EncodingRequest) (*EncodingResponse, error)
	mustEmbedUnimplementedEncoderBackendServer()
}

// UnimplementedEncoderBackendServer must be embedded to have forward compatible implementations.
type UnimplementedEncoderBackendServer struct {
}

func (UnimplementedEncoderBackendServer) Rot13(context.Context, *EncodingRequest) (*EncodingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Rot13 not implemented")
}
func (UnimplementedEncoderBackendServer) mustEmbedUnimplementedEncoderBackendServer() {}

// UnsafeEncoderBackendServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EncoderBackendServer will
// result in compilation errors.
type UnsafeEncoderBackendServer interface {
	mustEmbedUnimplementedEncoderBackendServer()
}

func RegisterEncoderBackendServer(s grpc.ServiceRegistrar, srv EncoderBackendServer) {
	s.RegisterService(&EncoderBackend_ServiceDesc, srv)
}

func _EncoderBackend_Rot13_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EncodingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EncoderBackendServer).Rot13(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/encoder_backend.EncoderBackend/Rot13",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EncoderBackendServer).Rot13(ctx, req.(*EncodingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// EncoderBackend_ServiceDesc is the grpc.ServiceDesc for EncoderBackend service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EncoderBackend_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "encoder_backend.EncoderBackend",
	HandlerType: (*EncoderBackendServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Rot13",
			Handler:    _EncoderBackend_Rot13_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "encoder_backend.proto",
}
