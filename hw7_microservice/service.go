package main

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"log"
	"net"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type bizServer struct{}

type adminServer struct {
	ctx context.Context

	//broadcastLogCh   chan *Event
	//addLogListenerCh chan chan *Event
	//logListeners     []chan *Event
	//
	//broadcastStatCh   chan *Stat
	//addStatListenerCh chan chan *Stat
	//statListeners     []chan *Stat
}

type server struct {
	acl map[string][]string
	adminServer
	bizServer
}

func StartMyMicroservice(ctx context.Context, listenAddr, ACLData string) (err error) {

	acl := make(map[string][]string)
	if err = json.Unmarshal([]byte(ACLData), &acl); err != nil {
		return
	}

	server := &server{}
	server.ctx = ctx

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

	return nil
}

func (s *server) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	return nil, nil
}

func (s *server) streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	return nil
}

func (s *adminServer) Logging(nothing *Nothing, srv Admin_LoggingServer) error {

	return nil
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
