package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"math"
	"net"
	"strings"
	"time"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type bizServer struct{}

type adminServer struct {
	ctx         context.Context
	logChan     chan *Event
	statChan    chan *Stat
	newLogChan  chan chan *Event
	listenLogs  []chan *Event
	newStatChan chan chan *Stat
	listenStats []chan *Stat
}

type server struct {
	acl map[string][]string
	adminServer
	bizServer
}

func StartMyMicroservice(ctx context.Context, listenAddr, ACLData string) (err error) {

	server := &server{}
	server.ctx = ctx

	if err = json.Unmarshal([]byte(ACLData), &server.acl); err != nil {
		return
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(server.streamInterceptor),
		grpc.UnaryInterceptor(server.unaryInterceptor),
	)

	RegisterAdminServer(grpcServer, server)
	RegisterBizServer(grpcServer, server)

	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		<-ctx.Done()
		//grpcServer.Stop()
		grpcServer.GracefulStop()
	}()

	server.logChan = make(chan *Event, 0)
	server.newLogChan = make(chan chan *Event, 0)

	go func() {
		for {
			select {
			case ch := <-server.newLogChan:
				server.listenLogs = append(server.listenLogs, ch)
			case event := <-server.logChan:
				for _, ch := range server.listenLogs {
					ch <- event
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	server.statChan = make(chan *Stat, 0)
	server.newStatChan = make(chan chan *Stat, 0)

	go func() {
		for {
			select {
			case ch := <-server.newStatChan:
				server.listenStats = append(server.listenStats, ch)
			case stat := <-server.statChan:
				for _, ch := range server.listenStats {
					ch <- stat
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (s *server) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	if err := s.checkAuth(ctx, info.FullMethod); err != nil {
		return "", err
	}

	err := s.parseMetadata(ctx, nil, info)
	if err != nil {
		return "", err
	}

	return handler(ctx, req)
}

func (s *server) streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	if err := s.checkAuth(ss.Context(), info.FullMethod); err != nil {
		return err
	}

	err := s.parseMetadata(ss.Context(), info, nil)
	if err != nil {
		return err
	}

	return handler(srv, ss)
}

func (s *server) parseMetadata(ctx context.Context, infoStream *grpc.StreamServerInfo, infoUnar *grpc.UnaryServerInfo) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "can't get metadata from incoming context")
	}

	consumer := md.Get("consumer")
	if len(consumer) != 1 {
		return status.Errorf(codes.Unauthenticated, "can't get consumer from metadata")
	}

	authority := md.Get(":authority")
	if len(authority) != 1 {
		return errors.New("can't get client host from metadata")
	}

	info := ""
	if infoStream != nil {
		info = infoStream.FullMethod
	} else {
		info = infoUnar.FullMethod
	}

	s.logChan <- &Event{
		Consumer: consumer[0],
		Method:   info,
		Host:     authority[0] + ":",
	}

	s.statChan <- &Stat{
		ByConsumer: map[string]uint64{consumer[0]: 1},
		ByMethod:   map[string]uint64{info: 1},
	}
	return nil
}

func (s *adminServer) Logging(nothing *Nothing, logServer Admin_LoggingServer) error {

	ch := make(chan *Event, 0)
	s.newLogChan <- ch

	for {
		select {
		case event := <-ch:
			err := logServer.Send(event)
			if err != nil {
				log.Println("error in send logs")
			}
		case <-s.ctx.Done():
			return nil
		}
	}
}

func (s *adminServer) Statistics(interval *StatInterval, statServer Admin_StatisticsServer) error {

	ch := make(chan *Stat, 0)
	s.newStatChan <- ch

	tick := time.NewTicker(time.Second * time.Duration(interval.IntervalSeconds))

	sum := &Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}

	for {
		select {
		case stat := <-ch:
			for key, val := range stat.ByMethod {
				sum.ByMethod[key] += val
			}
			for key, val := range stat.ByConsumer {
				sum.ByConsumer[key] += val
			}

		case <-tick.C:
			err := statServer.Send(sum)
			if err != nil {
				log.Println("error in send logs")
			}
			sum = &Stat{
				ByMethod:   make(map[string]uint64),
				ByConsumer: make(map[string]uint64),
			}

		case <-s.ctx.Done():
			return nil
		}
	}
}

func (bs *bizServer) Check(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (bs *bizServer) Add(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (bs *bizServer) Test(ctx context.Context, n *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (s *server) checkAuth(ctx context.Context, fullMethod string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "can't get metadata from incoming context")
	}

	consumer := md.Get("consumer")
	if len(consumer) != 1 {
		return status.Errorf(codes.Unauthenticated, "can't get consumer from metadata")
	}

	allowed, ok := s.acl[consumer[0]]
	if !ok {
		return status.Errorf(codes.Unauthenticated, "can't allow to enter 1")
	}

	methods := strings.Split(fullMethod, "/")
	if len(methods) != 3 {
		return status.Errorf(codes.Unauthenticated, "can't allow to enter 2")
	}

	path, method := methods[1], methods[2]
	isAuthed := false

	for _, allow := range allowed {
		s := strings.Split(allow, "/")
		if len(s) != 3 {
			continue
		}
		pathAllow, methodAllow := s[1], s[2]
		if path != pathAllow {
			continue
		}
		if methodAllow == "*" || method == methodAllow {
			isAuthed = true
			break
		}
	}

	if !isAuthed {
		return status.Errorf(codes.Unauthenticated, "can't allow to enter 3")
	}

	return nil
}

// ----------------------------------

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Event struct {
	Timestamp            int64    `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Consumer             string   `protobuf:"bytes,2,opt,name=consumer,proto3" json:"consumer,omitempty"`
	Method               string   `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	Host                 string   `protobuf:"bytes,4,opt,name=host,proto3" json:"host,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Event) Reset()         { *m = Event{} }
func (m *Event) String() string { return proto.CompactTextString(m) }
func (*Event) ProtoMessage()    {}
func (*Event) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_8108dcf1dd6080ef, []int{0}
}
func (m *Event) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Event.Unmarshal(m, b)
}
func (m *Event) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Event.Marshal(b, m, deterministic)
}
func (dst *Event) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Event.Merge(dst, src)
}
func (m *Event) XXX_Size() int {
	return xxx_messageInfo_Event.Size(m)
}
func (m *Event) XXX_DiscardUnknown() {
	xxx_messageInfo_Event.DiscardUnknown(m)
}

var xxx_messageInfo_Event proto.InternalMessageInfo

func (m *Event) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *Event) GetConsumer() string {
	if m != nil {
		return m.Consumer
	}
	return ""
}

func (m *Event) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *Event) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

type Stat struct {
	Timestamp            int64             `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	ByMethod             map[string]uint64 `protobuf:"bytes,2,rep,name=by_method,json=byMethod,proto3" json:"by_method,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	ByConsumer           map[string]uint64 `protobuf:"bytes,3,rep,name=by_consumer,json=byConsumer,proto3" json:"by_consumer,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Stat) Reset()         { *m = Stat{} }
func (m *Stat) String() string { return proto.CompactTextString(m) }
func (*Stat) ProtoMessage()    {}
func (*Stat) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_8108dcf1dd6080ef, []int{1}
}
func (m *Stat) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Stat.Unmarshal(m, b)
}
func (m *Stat) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Stat.Marshal(b, m, deterministic)
}
func (dst *Stat) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Stat.Merge(dst, src)
}
func (m *Stat) XXX_Size() int {
	return xxx_messageInfo_Stat.Size(m)
}
func (m *Stat) XXX_DiscardUnknown() {
	xxx_messageInfo_Stat.DiscardUnknown(m)
}

var xxx_messageInfo_Stat proto.InternalMessageInfo

func (m *Stat) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *Stat) GetByMethod() map[string]uint64 {
	if m != nil {
		return m.ByMethod
	}
	return nil
}

func (m *Stat) GetByConsumer() map[string]uint64 {
	if m != nil {
		return m.ByConsumer
	}
	return nil
}

type StatInterval struct {
	IntervalSeconds      uint64   `protobuf:"varint,1,opt,name=interval_seconds,json=intervalSeconds,proto3" json:"interval_seconds,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StatInterval) Reset()         { *m = StatInterval{} }
func (m *StatInterval) String() string { return proto.CompactTextString(m) }
func (*StatInterval) ProtoMessage()    {}
func (*StatInterval) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_8108dcf1dd6080ef, []int{2}
}
func (m *StatInterval) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StatInterval.Unmarshal(m, b)
}
func (m *StatInterval) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StatInterval.Marshal(b, m, deterministic)
}
func (dst *StatInterval) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StatInterval.Merge(dst, src)
}
func (m *StatInterval) XXX_Size() int {
	return xxx_messageInfo_StatInterval.Size(m)
}
func (m *StatInterval) XXX_DiscardUnknown() {
	xxx_messageInfo_StatInterval.DiscardUnknown(m)
}

var xxx_messageInfo_StatInterval proto.InternalMessageInfo

func (m *StatInterval) GetIntervalSeconds() uint64 {
	if m != nil {
		return m.IntervalSeconds
	}
	return 0
}

type Nothing struct {
	Dummy                bool     `protobuf:"varint,1,opt,name=dummy,proto3" json:"dummy,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Nothing) Reset()         { *m = Nothing{} }
func (m *Nothing) String() string { return proto.CompactTextString(m) }
func (*Nothing) ProtoMessage()    {}
func (*Nothing) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_8108dcf1dd6080ef, []int{3}
}
func (m *Nothing) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Nothing.Unmarshal(m, b)
}
func (m *Nothing) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Nothing.Marshal(b, m, deterministic)
}
func (dst *Nothing) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Nothing.Merge(dst, src)
}
func (m *Nothing) XXX_Size() int {
	return xxx_messageInfo_Nothing.Size(m)
}
func (m *Nothing) XXX_DiscardUnknown() {
	xxx_messageInfo_Nothing.DiscardUnknown(m)
}

var xxx_messageInfo_Nothing proto.InternalMessageInfo

func (m *Nothing) GetDummy() bool {
	if m != nil {
		return m.Dummy
	}
	return false
}

func init() {
	proto.RegisterType((*Event)(nil), "main.Event")
	proto.RegisterType((*Stat)(nil), "main.Stat")
	proto.RegisterMapType((map[string]uint64)(nil), "main.Stat.ByConsumerEntry")
	proto.RegisterMapType((map[string]uint64)(nil), "main.Stat.ByMethodEntry")
	proto.RegisterType((*StatInterval)(nil), "main.StatInterval")
	proto.RegisterType((*Nothing)(nil), "main.Nothing")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// AdminClient is the client API for Admin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AdminClient interface {
	Logging(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (Admin_LoggingClient, error)
	Statistics(ctx context.Context, in *StatInterval, opts ...grpc.CallOption) (Admin_StatisticsClient, error)
}

type adminClient struct {
	cc *grpc.ClientConn
}

func NewAdminClient(cc *grpc.ClientConn) AdminClient {
	return &adminClient{cc}
}

func (c *adminClient) Logging(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (Admin_LoggingClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Admin_serviceDesc.Streams[0], "/main.Admin/Logging", opts...)
	if err != nil {
		return nil, err
	}
	x := &adminLoggingClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Admin_LoggingClient interface {
	Recv() (*Event, error)
	grpc.ClientStream
}

type adminLoggingClient struct {
	grpc.ClientStream
}

func (x *adminLoggingClient) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *adminClient) Statistics(ctx context.Context, in *StatInterval, opts ...grpc.CallOption) (Admin_StatisticsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Admin_serviceDesc.Streams[1], "/main.Admin/Statistics", opts...)
	if err != nil {
		return nil, err
	}
	x := &adminStatisticsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Admin_StatisticsClient interface {
	Recv() (*Stat, error)
	grpc.ClientStream
}

type adminStatisticsClient struct {
	grpc.ClientStream
}

func (x *adminStatisticsClient) Recv() (*Stat, error) {
	m := new(Stat)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// AdminServer is the server API for Admin service.
type AdminServer interface {
	Logging(*Nothing, Admin_LoggingServer) error
	Statistics(*StatInterval, Admin_StatisticsServer) error
}

func RegisterAdminServer(s *grpc.Server, srv AdminServer) {
	s.RegisterService(&_Admin_serviceDesc, srv)
}

func _Admin_Logging_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Nothing)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AdminServer).Logging(m, &adminLoggingServer{stream})
}

type Admin_LoggingServer interface {
	Send(*Event) error
	grpc.ServerStream
}

type adminLoggingServer struct {
	grpc.ServerStream
}

func (x *adminLoggingServer) Send(m *Event) error {
	return x.ServerStream.SendMsg(m)
}

func _Admin_Statistics_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StatInterval)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AdminServer).Statistics(m, &adminStatisticsServer{stream})
}

type Admin_StatisticsServer interface {
	Send(*Stat) error
	grpc.ServerStream
}

type adminStatisticsServer struct {
	grpc.ServerStream
}

func (x *adminStatisticsServer) Send(m *Stat) error {
	return x.ServerStream.SendMsg(m)
}

var _Admin_serviceDesc = grpc.ServiceDesc{
	ServiceName: "main.Admin",
	HandlerType: (*AdminServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Logging",
			Handler:       _Admin_Logging_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Statistics",
			Handler:       _Admin_Statistics_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "service.proto",
}

// BizClient is the client API for Biz service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BizClient interface {
	Check(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*Nothing, error)
	Add(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*Nothing, error)
	Test(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*Nothing, error)
}

type bizClient struct {
	cc *grpc.ClientConn
}

func NewBizClient(cc *grpc.ClientConn) BizClient {
	return &bizClient{cc}
}

func (c *bizClient) Check(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/main.Biz/Check", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bizClient) Add(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/main.Biz/Add", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bizClient) Test(ctx context.Context, in *Nothing, opts ...grpc.CallOption) (*Nothing, error) {
	out := new(Nothing)
	err := c.cc.Invoke(ctx, "/main.Biz/Test", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BizServer is the server API for Biz service.
type BizServer interface {
	Check(context.Context, *Nothing) (*Nothing, error)
	Add(context.Context, *Nothing) (*Nothing, error)
	Test(context.Context, *Nothing) (*Nothing, error)
}

func RegisterBizServer(s *grpc.Server, srv BizServer) {
	s.RegisterService(&_Biz_serviceDesc, srv)
}

func _Biz_Check_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Nothing)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BizServer).Check(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.Biz/Check",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BizServer).Check(ctx, req.(*Nothing))
	}
	return interceptor(ctx, in, info, handler)
}

func _Biz_Add_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Nothing)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BizServer).Add(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.Biz/Add",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BizServer).Add(ctx, req.(*Nothing))
	}
	return interceptor(ctx, in, info, handler)
}

func _Biz_Test_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Nothing)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BizServer).Test(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.Biz/Test",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BizServer).Test(ctx, req.(*Nothing))
	}
	return interceptor(ctx, in, info, handler)
}

var _Biz_serviceDesc = grpc.ServiceDesc{
	ServiceName: "main.Biz",
	HandlerType: (*BizServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Check",
			Handler:    _Biz_Check_Handler,
		},
		{
			MethodName: "Add",
			Handler:    _Biz_Add_Handler,
		},
		{
			MethodName: "Test",
			Handler:    _Biz_Test_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "service.proto",
}

func init() { proto.RegisterFile("service.proto", fileDescriptor_service_8108dcf1dd6080ef) }

var fileDescriptor_service_8108dcf1dd6080ef = []byte{
	// 385 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x94, 0x52, 0x4d, 0x6b, 0xea, 0x40,
	0x14, 0x35, 0x5f, 0x6a, 0xae, 0x4f, 0x94, 0xe1, 0xf1, 0x08, 0xe1, 0xc1, 0x93, 0xc0, 0x6b, 0xed,
	0x26, 0x88, 0xa5, 0xd0, 0x56, 0xba, 0x50, 0x71, 0x51, 0x68, 0xbb, 0x88, 0xdd, 0x4b, 0x3e, 0x06,
	0x33, 0xe8, 0x24, 0x92, 0x19, 0x85, 0x14, 0xfa, 0x2f, 0xfa, 0x83, 0x9b, 0x99, 0x44, 0x45, 0x37,
	0xd2, 0xdd, 0x9c, 0x73, 0xef, 0x39, 0xf7, 0xe4, 0xe6, 0x42, 0x9b, 0xe1, 0x6c, 0x47, 0x42, 0xec,
	0x6e, 0xb2, 0x94, 0xa7, 0x48, 0xa7, 0x3e, 0x49, 0x1c, 0x0a, 0xc6, 0x6c, 0x87, 0x13, 0x8e, 0xfe,
	0x82, 0xc9, 0x09, 0xc5, 0x8c, 0xfb, 0x74, 0x63, 0x29, 0x3d, 0xa5, 0xaf, 0x79, 0x47, 0x02, 0xd9,
	0xd0, 0x0c, 0xd3, 0x84, 0x6d, 0x29, 0xce, 0x2c, 0xb5, 0x28, 0x9a, 0xde, 0x01, 0xa3, 0x3f, 0x50,
	0xa7, 0x98, 0xc7, 0x69, 0x64, 0x69, 0xb2, 0x52, 0x21, 0x84, 0x40, 0x8f, 0x53, 0xc6, 0x2d, 0x5d,
	0xb2, 0xf2, 0xed, 0x7c, 0xa9, 0xa0, 0xcf, 0xb9, 0x7f, 0x69, 0xdc, 0x1d, 0x98, 0x41, 0xbe, 0xa8,
	0x5c, 0xd5, 0x9e, 0xd6, 0x6f, 0x0d, 0x2d, 0x57, 0xe4, 0x75, 0x85, 0xd8, 0x9d, 0xe4, 0xaf, 0xb2,
	0x34, 0x4b, 0x78, 0x96, 0x7b, 0xcd, 0xa0, 0x82, 0x68, 0x04, 0xad, 0x42, 0x76, 0x08, 0xaa, 0x49,
	0xa1, 0x7d, 0x22, 0x9c, 0x56, 0xc5, 0x52, 0x0a, 0xc1, 0x81, 0xb0, 0x47, 0xd0, 0x3e, 0xf1, 0x45,
	0x5d, 0xd0, 0x56, 0x38, 0x97, 0xe1, 0x4c, 0x4f, 0x3c, 0xd1, 0x6f, 0x30, 0x76, 0xfe, 0x7a, 0x8b,
	0xe5, 0x0a, 0x74, 0xaf, 0x04, 0x8f, 0xea, 0xbd, 0x62, 0x3f, 0x41, 0xe7, 0xcc, 0xfb, 0x27, 0x72,
	0xe7, 0x01, 0x7e, 0x89, 0x7c, 0xcf, 0x09, 0x2f, 0x7e, 0x91, 0xbf, 0x46, 0x37, 0xd0, 0x25, 0xd5,
	0x7b, 0xc1, 0x70, 0xf1, 0x41, 0x11, 0x93, 0x46, 0xba, 0xd7, 0xd9, 0xf3, 0xf3, 0x92, 0x76, 0xfe,
	0x41, 0xe3, 0x2d, 0xe5, 0x31, 0x49, 0x96, 0xc2, 0x3f, 0xda, 0x52, 0x5a, 0xce, 0x6c, 0x7a, 0x25,
	0x18, 0x46, 0x60, 0x8c, 0x23, 0x4a, 0x92, 0xc2, 0xb4, 0xf1, 0x92, 0x2e, 0x97, 0xa2, 0xb3, 0x5d,
	0xee, 0xa4, 0x12, 0xda, 0xad, 0x12, 0xca, 0x43, 0x70, 0x6a, 0x03, 0x05, 0x0d, 0x00, 0x44, 0x1e,
	0xc2, 0x38, 0x09, 0x19, 0x42, 0xc7, 0x0d, 0xee, 0x13, 0xda, 0x70, 0xe4, 0x84, 0x62, 0xf8, 0x09,
	0xda, 0x84, 0x7c, 0xa0, 0x6b, 0x30, 0xa6, 0x31, 0x0e, 0x57, 0xe7, 0x13, 0x4e, 0xa1, 0x53, 0x43,
	0xff, 0x41, 0x1b, 0x47, 0xd1, 0xc5, 0xb6, 0x2b, 0xd0, 0xdf, 0x8b, 0x9b, 0xb8, 0xd4, 0x17, 0xd4,
	0xe5, 0x4d, 0xdf, 0x7e, 0x07, 0x00, 0x00, 0xff, 0xff, 0x03, 0x1d, 0xb2, 0x19, 0xe4, 0x02, 0x00,
	0x00,
}
