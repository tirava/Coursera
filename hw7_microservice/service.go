package main

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type bizServer struct{}

type adminServer struct {
	ctx        context.Context
	logChan    chan *Event
	statChan   chan *Stat
	newLogChan chan chan *Event
	listenLogs []chan *Event
	//
	//addStatListenerCh chan chan *Stat
	//statListeners     []chan *Stat
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

	return nil
}

func (s *server) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	if err := s.checkAuth(ctx, info.FullMethod); err != nil {
		return "", err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "can't get metadata from incoming context")
	}

	consumer := md.Get("consumer")
	if len(consumer) != 1 {
		return "", status.Errorf(codes.Unauthenticated, "can't get consumer from metadata")
	}

	s.logChan <- &Event{
		Consumer: consumer[0],
		Method:   info.FullMethod,
		Host:     "127.0.0.1:8083",
	}
	//s.statChan <- &Stat{
	//	ByConsumer: map[string]uint64{consumer[0]: 1},
	//	ByMethod:   map[string]uint64{info.FullMethod: 1},
	//}

	return handler(ctx, req)
}

func (s *server) streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	if err := s.checkAuth(ss.Context(), info.FullMethod); err != nil {
		return err
	}

	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Errorf(codes.Unauthenticated, "can't get metadata from incoming context")
	}

	consumer := md.Get("consumer")
	if len(consumer) != 1 {
		return status.Errorf(codes.Unauthenticated, "can't get consumer from metadata")
	}

	s.logChan <- &Event{
		Consumer: consumer[0],
		Method:   info.FullMethod,
		Host:     "127.0.0.1:8083",
	}
	//s.statChan <- &Stat{
	//	ByConsumer: map[string]uint64{consumer[0]: 1},
	//	ByMethod:   map[string]uint64{info.FullMethod: 1},
	//}

	return handler(srv, ss)
}

//func (s *server) getMetaData (ctx context.Context) consumerToDo  {
//
//}

func (s *adminServer) Logging(nothing *Nothing, logServer Admin_LoggingServer) error {

	ch := make(chan *Event, 0)
	s.newLogChan <- ch

	for {
		select {
		case event := <-ch:
			logServer.Send(event)
		case <-s.ctx.Done():
			return nil
		}
	}
}

func (s *adminServer) Statistics(interval *StatInterval, srv Admin_StatisticsServer) error {

	return nil
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
