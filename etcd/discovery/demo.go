package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/codetodo-io/storage-bootcamp/etcd/discovery/service"
)

var (
	dialTimeout  = 5 * time.Second
	leaseTimeout = int64(3)
	etcdPrefix   = "/etcd/discovery"
)

func runWorker() {
	var endpoints = []string{"etcd:2379"}
	ip, err := service.GetLocalIP()
	if err != nil {
		log.Fatal(err)
		return
	}
	workerInfo := &service.WorkerInfo{
		IP:   ip,
		Port: 8888,
	}
	worker, err := service.NewWorker(endpoints, *workerInfo, dialTimeout, etcdPrefix, leaseTimeout)
	if err != nil {
		log.Fatal("New worker Register error: ", err)
	}

	// 响应监听
	go worker.ListenRespC()
	select {
	// case <-time.After(20 * time.Second):
	// 	worker.Close()
	}
}

func runMaster() {
	var endpoints = []string{"etcd:2379"}
	server, err := service.NewMaster(endpoints, dialTimeout)
	if err != nil {
		log.Fatal("New Master Register error: ", err)
	}
	defer server.Close()

	server.WatchServer(etcdPrefix)

	for {
		select {
		case <-time.Tick(10 * time.Second):
			log.Println("workerList: ", server.GetWorkerList())
		}
	}
}

func main() {
	var role = flag.String("role", "", "--role=master | worker")
	flag.Parse()
	if *role == "master" {
		runMaster()
	} else if *role == "worker" {
		runWorker()
	} else {
		fmt.Println("demo -h for usage")
	}
}
