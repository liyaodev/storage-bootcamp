package main

import (
	"context"
	"log"
	"net"

	"github.com/codetodo-io/storage-bootcamp/etcd/lb/proto"
	"github.com/codetodo-io/storage-bootcamp/etcd/lb/service"
	"google.golang.org/grpc"
)

type HelloService struct{}

const (
	Address string = "127.0.0.1:8888"
	Network string = "tcp"
	Name    string = "hello_lb"
)

var endpoints = []string{"etcd:2379"}

func (s *HelloService) Route(ctx context.Context, req *proto.HelloRequest) (*proto.HelloResponse, error) {
	log.Println("receive: " + req.Data)
	res := proto.HelloResponse{
		Code:  200,
		Value: "Hello " + req.Data,
	}
	return &res, nil
}

func main() {
	l, err := net.Listen(Network, Address)
	if err != nil {
		log.Fatal("net.Listen err: %v", err)
	}
	log.Println(Address + " net.Listen...")

	server := grpc.NewServer()
	proto.RegisterHelloServer(server, &HelloService{})
	service, err := service.NewServiceRegister(endpoints, Name, Address, 5)
	if err != nil {
		log.Fatal("register service err: %v", err)
	}
	defer service.Close()

	err = server.Serve(l)
	if err != nil {
		log.Fatal("server.Serve err: %v", err)
	}
}
