package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/codetodo-io/storage-bootcamp/etcd/lb/proto"
	"github.com/codetodo-io/storage-bootcamp/etcd/lb/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

var (
	endpoints = []string{"etcd:2379"}
	name      = "hello_lb"
	client    = proto.HelloClient
)

func route(i int) {
	req := proto.HelloRequest{
		Data: "Hello " + strconv.Itoa(i),
	}
	res, err := client.Route(context.Background(), &req)
	if err != nil {
		log.Fatal("Call route err: %v", err)
	}
	log.Println(res)
}

func main() {
	r := service.NewServiceDiscovery(endpoints)
	resolver.Register(r)

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:///%s", r.Scheme(), name),
		grpc.WithBalancerName("round_robin"),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal("net.Connect err: %v", err)
	}
	defer conn.Close()

	client = proto.NewHelloClient(conn)
	for i := 0; i < 100; i++ {
		route(i)
		time.Sleep(1 * time.Second)
	}
}
