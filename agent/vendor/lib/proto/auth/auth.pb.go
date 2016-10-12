// Code generated by protoc-gen-go.
// source: auth.proto
// DO NOT EDIT!

/*
Package auth is a generated protocol buffer package.

It is generated from these files:
	auth.proto

It has these top-level messages:
	Auth
*/
package auth

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

type Auth_CertificateType int32

const (
	Auth_UUID     Auth_CertificateType = 0
	Auth_PLAIN    Auth_CertificateType = 1
	Auth_TOKEN    Auth_CertificateType = 2
	Auth_FACEBOOK Auth_CertificateType = 3
)

var Auth_CertificateType_name = map[int32]string{
	0: "UUID",
	1: "PLAIN",
	2: "TOKEN",
	3: "FACEBOOK",
}
var Auth_CertificateType_value = map[string]int32{
	"UUID":     0,
	"PLAIN":    1,
	"TOKEN":    2,
	"FACEBOOK": 3,
}

func (x Auth_CertificateType) String() string {
	return proto.EnumName(Auth_CertificateType_name, int32(x))
}
func (Auth_CertificateType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type Auth struct {
}

func (m *Auth) Reset()                    { *m = Auth{} }
func (m *Auth) String() string            { return proto.CompactTextString(m) }
func (*Auth) ProtoMessage()               {}
func (*Auth) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Auth_Certificate struct {
	Type  Auth_CertificateType `protobuf:"varint,1,opt,name=Type,json=type,enum=auth.Auth_CertificateType" json:"Type,omitempty"`
	Proof []byte               `protobuf:"bytes,2,opt,name=Proof,json=proof,proto3" json:"Proof,omitempty"`
}

func (m *Auth_Certificate) Reset()                    { *m = Auth_Certificate{} }
func (m *Auth_Certificate) String() string            { return proto.CompactTextString(m) }
func (*Auth_Certificate) ProtoMessage()               {}
func (*Auth_Certificate) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type Auth_Result struct {
	OK     bool   `protobuf:"varint,1,opt,name=OK,json=oK" json:"OK,omitempty"`
	UserId uint64 `protobuf:"varint,2,opt,name=UserId,json=userId" json:"UserId,omitempty"`
	Body   []byte `protobuf:"bytes,3,opt,name=Body,json=body,proto3" json:"Body,omitempty"`
}

func (m *Auth_Result) Reset()                    { *m = Auth_Result{} }
func (m *Auth_Result) String() string            { return proto.CompactTextString(m) }
func (*Auth_Result) ProtoMessage()               {}
func (*Auth_Result) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 1} }

func init() {
	proto.RegisterType((*Auth)(nil), "auth.Auth")
	proto.RegisterType((*Auth_Certificate)(nil), "auth.Auth.Certificate")
	proto.RegisterType((*Auth_Result)(nil), "auth.Auth.Result")
	proto.RegisterEnum("auth.Auth_CertificateType", Auth_CertificateType_name, Auth_CertificateType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for AuthService service

type AuthServiceClient interface {
	Auth(ctx context.Context, in *Auth_Certificate, opts ...grpc.CallOption) (*Auth_Result, error)
}

type authServiceClient struct {
	cc *grpc.ClientConn
}

func NewAuthServiceClient(cc *grpc.ClientConn) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) Auth(ctx context.Context, in *Auth_Certificate, opts ...grpc.CallOption) (*Auth_Result, error) {
	out := new(Auth_Result)
	err := grpc.Invoke(ctx, "/auth.AuthService/Auth", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for AuthService service

type AuthServiceServer interface {
	Auth(context.Context, *Auth_Certificate) (*Auth_Result, error)
}

func RegisterAuthServiceServer(s *grpc.Server, srv AuthServiceServer) {
	s.RegisterService(&_AuthService_serviceDesc, srv)
}

func _AuthService_Auth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Auth_Certificate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Auth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.AuthService/Auth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Auth(ctx, req.(*Auth_Certificate))
	}
	return interceptor(ctx, in, info, handler)
}

var _AuthService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "auth.AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Auth",
			Handler:    _AuthService_Auth_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("auth.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 252 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x50, 0xcd, 0x6a, 0xc2, 0x40,
	0x18, 0xec, 0xc6, 0x4d, 0x48, 0x3f, 0xc5, 0xa6, 0x1f, 0x45, 0x42, 0x4e, 0xe2, 0xc9, 0x53, 0xa0,
	0xf6, 0x01, 0xda, 0xf8, 0x53, 0x08, 0x29, 0x46, 0x56, 0xf3, 0x00, 0x6a, 0x3e, 0x31, 0x50, 0xd8,
	0xb0, 0x6e, 0x0a, 0xfb, 0xba, 0x7d, 0x92, 0xb2, 0xeb, 0x45, 0x8a, 0xb7, 0x99, 0x65, 0x66, 0x76,
	0xe6, 0x03, 0xd8, 0x77, 0xfa, 0x9c, 0xb6, 0x4a, 0x6a, 0x89, 0xdc, 0xe2, 0xc9, 0x2f, 0x03, 0x9e,
	0x75, 0xfa, 0x9c, 0x6c, 0xa1, 0xbf, 0x20, 0xa5, 0x9b, 0x53, 0x73, 0xdc, 0x6b, 0xc2, 0x14, 0xf8,
	0xce, 0xb4, 0x14, 0xb3, 0x31, 0x9b, 0x0e, 0x67, 0x49, 0xea, 0x8c, 0x56, 0x98, 0xde, 0xa8, 0xac,
	0x42, 0x70, 0x6d, 0x5a, 0xc2, 0x17, 0xf0, 0x37, 0x4a, 0xca, 0x53, 0xec, 0x8d, 0xd9, 0x74, 0x20,
	0xfc, 0xd6, 0x92, 0x64, 0x09, 0x81, 0xa0, 0x4b, 0xf7, 0xad, 0x71, 0x08, 0x5e, 0x59, 0xb8, 0xb4,
	0x50, 0x78, 0xb2, 0xc0, 0x11, 0x04, 0xd5, 0x85, 0x54, 0x5e, 0x3b, 0x03, 0x17, 0x41, 0xe7, 0x18,
	0x22, 0xf0, 0xb9, 0xac, 0x4d, 0xdc, 0x73, 0x31, 0xfc, 0x20, 0x6b, 0x33, 0x79, 0x87, 0xa7, 0x7f,
	0x9f, 0x62, 0x08, 0xbc, 0xaa, 0xf2, 0x65, 0xf4, 0x80, 0x8f, 0xe0, 0x6f, 0xbe, 0xb2, 0x7c, 0x1d,
	0x31, 0x0b, 0x77, 0x65, 0xb1, 0x5a, 0x47, 0x1e, 0x0e, 0x20, 0xfc, 0xcc, 0x16, 0xab, 0x79, 0x59,
	0x16, 0x51, 0x6f, 0xf6, 0x01, 0x7d, 0x5b, 0x7d, 0x4b, 0xea, 0xa7, 0x39, 0x12, 0xbe, 0x5e, 0x27,
	0xe3, 0xe8, 0xfe, 0xaa, 0xe4, 0xf9, 0xe6, 0xfd, 0x5a, 0xff, 0x10, 0xb8, 0x9b, 0xbd, 0xfd, 0x05,
	0x00, 0x00, 0xff, 0xff, 0x30, 0x14, 0xef, 0xf0, 0x41, 0x01, 0x00, 0x00,
}
