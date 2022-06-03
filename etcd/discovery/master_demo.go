package main

import (
	"log"
	"time"

	"github.com/codetodo-io/storage-bootcamp/etcd/discovery/service"
)

var (
	etcdPrefix = "/etcd/discovery"
)

func main() {
	var endpoints = []string{"etcd:2379"}
	server, err := service.NewMaster(endpoints)
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
