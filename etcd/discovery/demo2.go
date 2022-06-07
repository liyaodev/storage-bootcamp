package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/codetodo-io/storage-bootcamp/etcd/discovery/service"
)

var (
	endpoints = []string{"etcd:2379"}
)

func register() {
	s, err := service.NewServiceRegister(endpoints, "/etcd/demo", "localhost:8000", 5)
	if err != nil {
		log.Println(err)
	}
	go s.ListenLeaseRespC()
	select {
	// case <-time.After(20 * time.Second):
	// 	s.Close()
	}
}

func discovery() {
	s := service.NewServiceDiscovery(endpoints)
	defer s.Close()
	s.WatchService("/etcd/")
	for {
		select {
		case <-time.Tick(10 * time.Second):
			log.Println(s.GetServices())
		}
	}
}

func main() {
	var role = flag.String("role", "", "--role=discovery | register")
	flag.Parse()
	if *role == "discovery" {
		discovery()
	} else if *role == "register" {
		register()
	} else {
		fmt.Println("demo -h for usage")
	}
}
