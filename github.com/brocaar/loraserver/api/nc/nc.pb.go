// Code generated by protoc-gen-go. DO NOT EDIT.
// source: nc.proto

package nc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import gw "github.com/brocaar/loraserver/api/gw"
import empty "github.com/golang/protobuf/ptypes/empty"

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

type HandleUplinkMetaDataRequest struct {
	// Device EUI (8 bytes).
	DevEui []byte `protobuf:"bytes,1,opt,name=dev_eui,json=devEui,proto3" json:"dev_eui,omitempty"`
	// TX meta-data.
	TxInfo *gw.UplinkTXInfo `protobuf:"bytes,2,opt,name=tx_info,json=txInfo,proto3" json:"tx_info,omitempty"`
	// RX meta-data.
	RxInfo               []*gw.UplinkRXInfo `protobuf:"bytes,3,rep,name=rx_info,json=rxInfo,proto3" json:"rx_info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *HandleUplinkMetaDataRequest) Reset()         { *m = HandleUplinkMetaDataRequest{} }
func (m *HandleUplinkMetaDataRequest) String() string { return proto.CompactTextString(m) }
func (*HandleUplinkMetaDataRequest) ProtoMessage()    {}
func (*HandleUplinkMetaDataRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fd7d898ee948e33, []int{0}
}
func (m *HandleUplinkMetaDataRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HandleUplinkMetaDataRequest.Unmarshal(m, b)
}
func (m *HandleUplinkMetaDataRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HandleUplinkMetaDataRequest.Marshal(b, m, deterministic)
}
func (dst *HandleUplinkMetaDataRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HandleUplinkMetaDataRequest.Merge(dst, src)
}
func (m *HandleUplinkMetaDataRequest) XXX_Size() int {
	return xxx_messageInfo_HandleUplinkMetaDataRequest.Size(m)
}
func (m *HandleUplinkMetaDataRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_HandleUplinkMetaDataRequest.DiscardUnknown(m)
}

var xxx_messageInfo_HandleUplinkMetaDataRequest proto.InternalMessageInfo

func (m *HandleUplinkMetaDataRequest) GetDevEui() []byte {
	if m != nil {
		return m.DevEui
	}
	return nil
}

func (m *HandleUplinkMetaDataRequest) GetTxInfo() *gw.UplinkTXInfo {
	if m != nil {
		return m.TxInfo
	}
	return nil
}

func (m *HandleUplinkMetaDataRequest) GetRxInfo() []*gw.UplinkRXInfo {
	if m != nil {
		return m.RxInfo
	}
	return nil
}

type HandleUplinkMACCommandRequest struct {
	// Device EUI (8 bytes).
	DevEui []byte `protobuf:"bytes,1,opt,name=dev_eui,json=devEui,proto3" json:"dev_eui,omitempty"`
	// Command identifier (specified by the LoRaWAN specs).
	Cid uint32 `protobuf:"varint,2,opt,name=cid,proto3" json:"cid,omitempty"`
	// MAC-command payload(s).
	Commands             [][]byte `protobuf:"bytes,6,rep,name=commands,proto3" json:"commands,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HandleUplinkMACCommandRequest) Reset()         { *m = HandleUplinkMACCommandRequest{} }
func (m *HandleUplinkMACCommandRequest) String() string { return proto.CompactTextString(m) }
func (*HandleUplinkMACCommandRequest) ProtoMessage()    {}
func (*HandleUplinkMACCommandRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3fd7d898ee948e33, []int{1}
}
func (m *HandleUplinkMACCommandRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HandleUplinkMACCommandRequest.Unmarshal(m, b)
}
func (m *HandleUplinkMACCommandRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HandleUplinkMACCommandRequest.Marshal(b, m, deterministic)
}
func (dst *HandleUplinkMACCommandRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HandleUplinkMACCommandRequest.Merge(dst, src)
}
func (m *HandleUplinkMACCommandRequest) XXX_Size() int {
	return xxx_messageInfo_HandleUplinkMACCommandRequest.Size(m)
}
func (m *HandleUplinkMACCommandRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_HandleUplinkMACCommandRequest.DiscardUnknown(m)
}

var xxx_messageInfo_HandleUplinkMACCommandRequest proto.InternalMessageInfo

func (m *HandleUplinkMACCommandRequest) GetDevEui() []byte {
	if m != nil {
		return m.DevEui
	}
	return nil
}

func (m *HandleUplinkMACCommandRequest) GetCid() uint32 {
	if m != nil {
		return m.Cid
	}
	return 0
}

func (m *HandleUplinkMACCommandRequest) GetCommands() [][]byte {
	if m != nil {
		return m.Commands
	}
	return nil
}

func init() {
	proto.RegisterType((*HandleUplinkMetaDataRequest)(nil), "nc.HandleUplinkMetaDataRequest")
	proto.RegisterType((*HandleUplinkMACCommandRequest)(nil), "nc.HandleUplinkMACCommandRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// NetworkControllerServiceClient is the client API for NetworkControllerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NetworkControllerServiceClient interface {
	// HandleUplinkMetaData handles uplink meta-rata.
	HandleUplinkMetaData(ctx context.Context, in *HandleUplinkMetaDataRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	// HandleUplinkMACCommand handles an uplink mac-command.
	// This method will only be called in case the mac-command request was
	// enqueued throught the API or when the CID is >= 0x80 (proprietary
	// mac-command range).
	HandleUplinkMACCommand(ctx context.Context, in *HandleUplinkMACCommandRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}

type networkControllerServiceClient struct {
	cc *grpc.ClientConn
}

func NewNetworkControllerServiceClient(cc *grpc.ClientConn) NetworkControllerServiceClient {
	return &networkControllerServiceClient{cc}
}

func (c *networkControllerServiceClient) HandleUplinkMetaData(ctx context.Context, in *HandleUplinkMetaDataRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/nc.NetworkControllerService/HandleUplinkMetaData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *networkControllerServiceClient) HandleUplinkMACCommand(ctx context.Context, in *HandleUplinkMACCommandRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/nc.NetworkControllerService/HandleUplinkMACCommand", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NetworkControllerServiceServer is the server API for NetworkControllerService service.
type NetworkControllerServiceServer interface {
	// HandleUplinkMetaData handles uplink meta-rata.
	HandleUplinkMetaData(context.Context, *HandleUplinkMetaDataRequest) (*empty.Empty, error)
	// HandleUplinkMACCommand handles an uplink mac-command.
	// This method will only be called in case the mac-command request was
	// enqueued throught the API or when the CID is >= 0x80 (proprietary
	// mac-command range).
	HandleUplinkMACCommand(context.Context, *HandleUplinkMACCommandRequest) (*empty.Empty, error)
}

func RegisterNetworkControllerServiceServer(s *grpc.Server, srv NetworkControllerServiceServer) {
	s.RegisterService(&_NetworkControllerService_serviceDesc, srv)
}

func _NetworkControllerService_HandleUplinkMetaData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandleUplinkMetaDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NetworkControllerServiceServer).HandleUplinkMetaData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/nc.NetworkControllerService/HandleUplinkMetaData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NetworkControllerServiceServer).HandleUplinkMetaData(ctx, req.(*HandleUplinkMetaDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NetworkControllerService_HandleUplinkMACCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HandleUplinkMACCommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NetworkControllerServiceServer).HandleUplinkMACCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/nc.NetworkControllerService/HandleUplinkMACCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NetworkControllerServiceServer).HandleUplinkMACCommand(ctx, req.(*HandleUplinkMACCommandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _NetworkControllerService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "nc.NetworkControllerService",
	HandlerType: (*NetworkControllerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HandleUplinkMetaData",
			Handler:    _NetworkControllerService_HandleUplinkMetaData_Handler,
		},
		{
			MethodName: "HandleUplinkMACCommand",
			Handler:    _NetworkControllerService_HandleUplinkMACCommand_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "nc.proto",
}

func init() { proto.RegisterFile("nc.proto", fileDescriptor_3fd7d898ee948e33) }

var fileDescriptor_3fd7d898ee948e33 = []byte{
	// 326 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0xd1, 0x4a, 0xeb, 0x30,
	0x1c, 0xc6, 0x4f, 0x57, 0xe8, 0x46, 0xce, 0x0e, 0x8c, 0x70, 0x98, 0xa5, 0x43, 0xac, 0xbb, 0x9a,
	0x17, 0x26, 0x30, 0x9f, 0x40, 0xe6, 0x40, 0x2f, 0x14, 0xac, 0x0e, 0xbc, 0x1b, 0x69, 0xfa, 0x6f,
	0x0d, 0x6b, 0x93, 0x9a, 0xa5, 0xed, 0x7c, 0x07, 0x1f, 0xcb, 0x07, 0x93, 0xb6, 0x53, 0x98, 0x4e,
	0xbd, 0x4a, 0xf2, 0xe7, 0x97, 0x8f, 0xef, 0xff, 0x7d, 0xa8, 0x27, 0x39, 0xc9, 0xb5, 0x32, 0x0a,
	0x77, 0x24, 0xf7, 0x46, 0x89, 0x52, 0x49, 0x0a, 0xb4, 0x99, 0x84, 0x45, 0x4c, 0x21, 0xcb, 0xcd,
	0x73, 0x0b, 0x78, 0xa7, 0x89, 0x30, 0x8f, 0x45, 0x48, 0xb8, 0xca, 0x68, 0xa8, 0x15, 0x67, 0x4c,
	0xd3, 0x54, 0x69, 0xb6, 0x06, 0x5d, 0x82, 0xa6, 0x2c, 0x17, 0x34, 0xa9, 0x68, 0x52, 0xb5, 0xf8,
	0xf8, 0xc5, 0x42, 0xa3, 0x4b, 0x26, 0xa3, 0x14, 0x16, 0x79, 0x2a, 0xe4, 0xea, 0x1a, 0x0c, 0xbb,
	0x60, 0x86, 0x05, 0xf0, 0x54, 0xc0, 0xda, 0xe0, 0x03, 0xd4, 0x8d, 0xa0, 0x5c, 0x42, 0x21, 0x5c,
	0xcb, 0xb7, 0x26, 0xfd, 0xc0, 0x89, 0xa0, 0x9c, 0x17, 0x02, 0x9f, 0xa0, 0xae, 0xd9, 0x2c, 0x85,
	0x8c, 0x95, 0xdb, 0xf1, 0xad, 0xc9, 0xdf, 0xe9, 0x80, 0x24, 0x15, 0x69, 0x45, 0xee, 0x1f, 0xae,
	0x64, 0xac, 0x02, 0xc7, 0x6c, 0xea, 0xb3, 0x46, 0xf5, 0x16, 0xb5, 0x7d, 0x7b, 0x17, 0x0d, 0xb6,
	0xa8, 0x6e, 0xd0, 0x71, 0x8c, 0x0e, 0x77, 0xdc, 0x9c, 0xcf, 0x66, 0x2a, 0xcb, 0x98, 0x8c, 0x7e,
	0xf5, 0x33, 0x40, 0x36, 0x17, 0x51, 0xe3, 0xe5, 0x5f, 0x50, 0x5f, 0xb1, 0x87, 0x7a, 0xbc, 0xfd,
	0xbc, 0x76, 0x1d, 0xdf, 0x9e, 0xf4, 0x83, 0x8f, 0xf7, 0xf4, 0xd5, 0x42, 0xee, 0x0d, 0x98, 0x4a,
	0xe9, 0xd5, 0x4c, 0x49, 0xa3, 0x55, 0x9a, 0x82, 0xbe, 0x03, 0x5d, 0x0a, 0x0e, 0xf8, 0x16, 0xfd,
	0xdf, 0x17, 0x09, 0x3e, 0x22, 0x92, 0x93, 0x1f, 0xc2, 0xf2, 0x86, 0xa4, 0x6d, 0x86, 0xbc, 0x37,
	0x43, 0xe6, 0x75, 0x33, 0xe3, 0x3f, 0x78, 0x81, 0x86, 0xfb, 0xf7, 0xc2, 0xc7, 0x5f, 0x44, 0x3f,
	0xef, 0xfc, 0xbd, 0x6c, 0xe8, 0x34, 0x93, 0xb3, 0xb7, 0x00, 0x00, 0x00, 0xff, 0xff, 0x9b, 0x82,
	0x2c, 0xe8, 0x20, 0x02, 0x00, 0x00,
}