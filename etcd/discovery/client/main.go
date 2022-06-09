package main

import (
	"log"
	"time"

	"github.com/codetodo-io/storage-bootcamp/etcd/discovery/service"
)

func main() {

	var endpoints = []string{"etcd:2379"}
	var etcdPrefix = "/etcd/discovery"

	service := service.NewServiceDiscovery(endpoints)
	defer service.Close()

	service.WatchService(etcdPrefix)

	for {
		select {
		case <-time.Tick(10 * time.Second):
			log.Println("serviceList: ", service.GetServices())
		}
	}
}
