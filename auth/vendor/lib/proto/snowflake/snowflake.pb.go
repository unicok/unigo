// Code generated by protoc-gen-go.
// source: snowflake.proto
// DO NOT EDIT!

/*
Package snowflake is a generated protocol buffer package.

It is generated from these files:
	snowflake.proto

It has these top-level messages:
	Snowflake
*/
package snowflake

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Snowflake struct {
}

func (m *Snowflake) Reset()                    { *m = Snowflake{} }
func (m *Snowflake) String() string            { return proto.CompactTextString(m) }
func (*Snowflake) ProtoMessage()               {}
func (*Snowflake) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Snowflake_Key struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *Snowflake_Key) Reset()                    { *m = Snowflake_Key{} }
func (m *Snowflake_Key) String() string            { return proto.CompactTextString(m) }
func (*Snowflake_Key) ProtoMessage()               {}
func (*Snowflake_Key) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type Snowflake_Value struct {
	Value int64 `protobuf:"varint,1,opt,name=value" json:"value,omitempty"`
}

func (m *Snowflake_Value) Reset()                    { *m = Snowflake_Value{} }
func (m *Snowflake_Value) String() string            { return proto.CompactTextString(m) }
func (*Snowflake_Value) ProtoMessage()               {}
func (*Snowflake_Value) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 1} }

type Snowflake_NullRequest struct {
}

func (m *Snowflake_NullRequest) Reset()                    { *m = Snowflake_NullRequest{} }
func (m *Snowflake_NullRequest) String() string            { return proto.CompactTextString(m) }
func (*Snowflake_NullRequest) ProtoMessage()               {}
func (*Snowflake_NullRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 2} }

type Snowflake_UUID struct {
	Uuid uint64 `protobuf:"varint,1,opt,name=uuid" json:"uuid,omitempty"`
}

func (m *Snowflake_UUID) Reset()                    { *m = Snowflake_UUID{} }
func (m *Snowflake_UUID) String() string            { return proto.CompactTextString(m) }
func (*Snowflake_UUID) ProtoMessage()               {}
func (*Snowflake_UUID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 3} }

func init() {
	proto.RegisterType((*Snowflake)(nil), "snowflake.Snowflake")
	proto.RegisterType((*Snowflake_Key)(nil), "snowflake.Snowflake.Key")
	proto.RegisterType((*Snowflake_Value)(nil), "snowflake.Snowflake.Value")
	proto.RegisterType((*Snowflake_NullRequest)(nil), "snowflake.Snowflake.NullRequest")
	proto.RegisterType((*Snowflake_UUID)(nil), "snowflake.Snowflake.UUID")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for SnowflakeService service

type SnowflakeServiceClient interface {
	Next(ctx context.Context, in *Snowflake_Key, opts ...grpc.CallOption) (*Snowflake_Value, error)
	GetUUID(ctx context.Context, in *Snowflake_NullRequest, opts ...grpc.CallOption) (*Snowflake_UUID, error)
}

type snowflakeServiceClient struct {
	cc *grpc.ClientConn
}

func NewSnowflakeServiceClient(cc *grpc.ClientConn) SnowflakeServiceClient {
	return &snowflakeServiceClient{cc}
}

func (c *snowflakeServiceClient) Next(ctx context.Context, in *Snowflake_Key, opts ...grpc.CallOption) (*Snowflake_Value, error) {
	out := new(Snowflake_Value)
	err := grpc.Invoke(ctx, "/snowflake.SnowflakeService/Next", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *snowflakeServiceClient) GetUUID(ctx context.Context, in *Snowflake_NullRequest, opts ...grpc.CallOption) (*Snowflake_UUID, error) {
	out := new(Snowflake_UUID)
	err := grpc.Invoke(ctx, "/snowflake.SnowflakeService/GetUUID", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for SnowflakeService service

type SnowflakeServiceServer interface {
	Next(context.Context, *Snowflake_Key) (*Snowflake_Value, error)
	GetUUID(context.Context, *Snowflake_NullRequest) (*Snowflake_UUID, error)
}

func RegisterSnowflakeServiceServer(s *grpc.Server, srv SnowflakeServiceServer) {
	s.RegisterService(&_SnowflakeService_serviceDesc, srv)
}

func _SnowflakeService_Next_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Snowflake_Key)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SnowflakeServiceServer).Next(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/snowflake.SnowflakeService/Next",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SnowflakeServiceServer).Next(ctx, req.(*Snowflake_Key))
	}
	return interceptor(ctx, in, info, handler)
}

func _SnowflakeService_GetUUID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Snowflake_NullRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SnowflakeServiceServer).GetUUID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/snowflake.SnowflakeService/GetUUID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SnowflakeServiceServer).GetUUID(ctx, req.(*Snowflake_NullRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SnowflakeService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "snowflake.SnowflakeService",
	HandlerType: (*SnowflakeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Next",
			Handler:    _SnowflakeService_Next_Handler,
		},
		{
			MethodName: "GetUUID",
			Handler:    _SnowflakeService_GetUUID_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("snowflake.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 196 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x2f, 0xce, 0xcb, 0x2f,
	0x4f, 0xcb, 0x49, 0xcc, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x84, 0x0b, 0x28,
	0x15, 0x70, 0x71, 0x06, 0xc3, 0x38, 0x52, 0x92, 0x5c, 0xcc, 0xde, 0xa9, 0x95, 0x42, 0x42, 0x5c,
	0x2c, 0x79, 0x89, 0xb9, 0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x60, 0xb6, 0x94, 0x2c,
	0x17, 0x6b, 0x58, 0x62, 0x4e, 0x69, 0xaa, 0x90, 0x08, 0x17, 0x6b, 0x19, 0x88, 0x01, 0x96, 0x65,
	0x0e, 0x82, 0x70, 0xa4, 0x78, 0xb9, 0xb8, 0xfd, 0x4a, 0x73, 0x72, 0x82, 0x52, 0x0b, 0x4b, 0x53,
	0x8b, 0x4b, 0xa4, 0xa4, 0xb8, 0x58, 0x42, 0x43, 0x3d, 0x5d, 0x40, 0x26, 0x95, 0x96, 0x66, 0xa6,
	0x80, 0xd5, 0xb2, 0x04, 0x81, 0xd9, 0x46, 0x33, 0x18, 0xb9, 0x04, 0xe0, 0x56, 0x06, 0xa7, 0x16,
	0x95, 0x65, 0x26, 0xa7, 0x0a, 0xd9, 0x70, 0xb1, 0xf8, 0xa5, 0x56, 0x94, 0x08, 0x49, 0xe8, 0x21,
	0xdc, 0x0a, 0x57, 0xa4, 0xe7, 0x9d, 0x5a, 0x29, 0x25, 0x85, 0x55, 0x06, 0xe2, 0x26, 0x37, 0x2e,
	0x76, 0xf7, 0xd4, 0x12, 0xb0, 0x8d, 0x0a, 0x58, 0x95, 0x21, 0xbb, 0x4d, 0x12, 0xab, 0x0a, 0x90,
	0xe6, 0x24, 0x36, 0x70, 0xf0, 0x18, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x32, 0x98, 0x60, 0x62,
	0x31, 0x01, 0x00, 0x00,
}
