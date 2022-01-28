// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package helloworld

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

// GreeterClient is the client API for Greeter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GreeterClient interface {
	SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error)
	SayHelloCS(ctx context.Context, opts ...grpc.CallOption) (Greeter_SayHelloCSClient, error)
	SayHellos(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (Greeter_SayHellosClient, error)
	SayHelloBidi(ctx context.Context, opts ...grpc.CallOption) (Greeter_SayHelloBidiClient, error)
}

type greeterClient struct {
	cc grpc.ClientConnInterface
}

func NewGreeterClient(cc grpc.ClientConnInterface) GreeterClient {
	return &greeterClient{cc}
}

func (c *greeterClient) SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error) {
	out := new(HelloReply)
	err := c.cc.Invoke(ctx, "/helloworld.Greeter/SayHello", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *greeterClient) SayHelloCS(ctx context.Context, opts ...grpc.CallOption) (Greeter_SayHelloCSClient, error) {
	stream, err := c.cc.NewStream(ctx, &Greeter_ServiceDesc.Streams[0], "/helloworld.Greeter/SayHelloCS", opts...)
	if err != nil {
		return nil, err
	}
	x := &greeterSayHelloCSClient{stream}
	return x, nil
}

type Greeter_SayHelloCSClient interface {
	Send(*HelloRequest) error
	CloseAndRecv() (*HelloReply, error)
	grpc.ClientStream
}

type greeterSayHelloCSClient struct {
	grpc.ClientStream
}

func (x *greeterSayHelloCSClient) Send(m *HelloRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *greeterSayHelloCSClient) CloseAndRecv() (*HelloReply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(HelloReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *greeterClient) SayHellos(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (Greeter_SayHellosClient, error) {
	stream, err := c.cc.NewStream(ctx, &Greeter_ServiceDesc.Streams[1], "/helloworld.Greeter/SayHellos", opts...)
	if err != nil {
		return nil, err
	}
	x := &greeterSayHellosClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Greeter_SayHellosClient interface {
	Recv() (*HelloReply, error)
	grpc.ClientStream
}

type greeterSayHellosClient struct {
	grpc.ClientStream
}

func (x *greeterSayHellosClient) Recv() (*HelloReply, error) {
	m := new(HelloReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *greeterClient) SayHelloBidi(ctx context.Context, opts ...grpc.CallOption) (Greeter_SayHelloBidiClient, error) {
	stream, err := c.cc.NewStream(ctx, &Greeter_ServiceDesc.Streams[2], "/helloworld.Greeter/SayHelloBidi", opts...)
	if err != nil {
		return nil, err
	}
	x := &greeterSayHelloBidiClient{stream}
	return x, nil
}

type Greeter_SayHelloBidiClient interface {
	Send(*HelloRequest) error
	Recv() (*HelloReply, error)
	grpc.ClientStream
}

type greeterSayHelloBidiClient struct {
	grpc.ClientStream
}

func (x *greeterSayHelloBidiClient) Send(m *HelloRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *greeterSayHelloBidiClient) Recv() (*HelloReply, error) {
	m := new(HelloReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// GreeterServer is the server API for Greeter service.
// All implementations must embed UnimplementedGreeterServer
// for forward compatibility
type GreeterServer interface {
	SayHello(context.Context, *HelloRequest) (*HelloReply, error)
	SayHelloCS(Greeter_SayHelloCSServer) error
	SayHellos(*HelloRequest, Greeter_SayHellosServer) error
	SayHelloBidi(Greeter_SayHelloBidiServer) error
	mustEmbedUnimplementedGreeterServer()
}

// UnimplementedGreeterServer must be embedded to have forward compatible implementations.
type UnimplementedGreeterServer struct {
}

func (UnimplementedGreeterServer) SayHello(context.Context, *HelloRequest) (*HelloReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SayHello not implemented")
}
func (UnimplementedGreeterServer) SayHelloCS(Greeter_SayHelloCSServer) error {
	return status.Errorf(codes.Unimplemented, "method SayHelloCS not implemented")
}
func (UnimplementedGreeterServer) SayHellos(*HelloRequest, Greeter_SayHellosServer) error {
	return status.Errorf(codes.Unimplemented, "method SayHellos not implemented")
}
func (UnimplementedGreeterServer) SayHelloBidi(Greeter_SayHelloBidiServer) error {
	return status.Errorf(codes.Unimplemented, "method SayHelloBidi not implemented")
}
func (UnimplementedGreeterServer) mustEmbedUnimplementedGreeterServer() {}

// UnsafeGreeterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GreeterServer will
// result in compilation errors.
type UnsafeGreeterServer interface {
	mustEmbedUnimplementedGreeterServer()
}

func RegisterGreeterServer(s grpc.ServiceRegistrar, srv GreeterServer) {
	s.RegisterService(&Greeter_ServiceDesc, srv)
}

func _Greeter_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/helloworld.Greeter/SayHello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).SayHello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Greeter_SayHelloCS_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GreeterServer).SayHelloCS(&greeterSayHelloCSServer{stream})
}

type Greeter_SayHelloCSServer interface {
	SendAndClose(*HelloReply) error
	Recv() (*HelloRequest, error)
	grpc.ServerStream
}

type greeterSayHelloCSServer struct {
	grpc.ServerStream
}

func (x *greeterSayHelloCSServer) SendAndClose(m *HelloReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *greeterSayHelloCSServer) Recv() (*HelloRequest, error) {
	m := new(HelloRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Greeter_SayHellos_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(HelloRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(GreeterServer).SayHellos(m, &greeterSayHellosServer{stream})
}

type Greeter_SayHellosServer interface {
	Send(*HelloReply) error
	grpc.ServerStream
}

type greeterSayHellosServer struct {
	grpc.ServerStream
}

func (x *greeterSayHellosServer) Send(m *HelloReply) error {
	return x.ServerStream.SendMsg(m)
}

func _Greeter_SayHelloBidi_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(GreeterServer).SayHelloBidi(&greeterSayHelloBidiServer{stream})
}

type Greeter_SayHelloBidiServer interface {
	Send(*HelloReply) error
	Recv() (*HelloRequest, error)
	grpc.ServerStream
}

type greeterSayHelloBidiServer struct {
	grpc.ServerStream
}

func (x *greeterSayHelloBidiServer) Send(m *HelloReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *greeterSayHelloBidiServer) Recv() (*HelloRequest, error) {
	m := new(HelloRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Greeter_ServiceDesc is the grpc.ServiceDesc for Greeter service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Greeter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "helloworld.Greeter",
	HandlerType: (*GreeterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _Greeter_SayHello_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SayHelloCS",
			Handler:       _Greeter_SayHelloCS_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "SayHellos",
			Handler:       _Greeter_SayHellos_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SayHelloBidi",
			Handler:       _Greeter_SayHelloBidi_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "helloworld/greeter.proto",
}
