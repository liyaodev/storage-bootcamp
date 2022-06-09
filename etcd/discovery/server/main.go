package main

import (
	"errors"
	"log"
	"net"
	"strconv"

	"github.com/codetodo-io/storage-bootcamp/etcd/discovery/service"
)

type NodeInfo struct {
	IP     string
	Port   int
	prefix string
}

func (node *NodeInfo) getKey() string {
	return node.prefix + "/" + node.IP + ":" + strconv.Itoa(node.Port)
}

func (node *NodeInfo) getVal() string {
	return "http://" + node.IP + ":" + strconv.Itoa(node.Port)
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// 检查 ip 地址，判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("Can not find the client ip address!")
}

func main() {
	var endpoints = []string{"etcd:2379"}
	var etcdPrefix = "/etcd/discovery"

	ip, err := getLocalIP()
	if err != nil {
		log.Fatal(err)
		return
	}
	node := &NodeInfo{
		IP:     ip,
		Port:   8888,
		prefix: etcdPrefix,
	}

	s, err := service.NewServiceRegister(endpoints, node.getKey(), node.getVal(), 5)
	if err != nil {
		log.Println(err)
	}

	// 响应监听
	go s.ListenLeaseRespC()
	select {
	// case <-time.After(20 * time.Second):
	// 	worker.Close()
	}
}
