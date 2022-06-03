package main

import (
	"log"

	"github.com/codetodo-io/storage-bootcamp/etcd/discovery/common"
)

func main() {
	var endpoints = []string{"etcd:2379"}
	ip, err := common.getLocalIP()
	if err != nil {
		log.Fatal(err)
		return
	}
	workerInfo := &common.WorkerInfo{
		IP:   ip,
		Port: 8888,
	}
	worker, err := common.NewWorker(endpoints, *workerInfo, common.leaseTimeout)
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
