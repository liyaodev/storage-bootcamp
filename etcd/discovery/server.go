package main

import (
	"context"
	"errors"
	"log"
	"net"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var (
	dialTimeout  = 5 * time.Second
	leaseTimeout = int64(3)
	etcdPrefix   = "/etcd/discovery"
)

type ServerInfo struct {
	ip   string
	port int
}

type ServerRegister struct {
	cli        *clientv3.Client
	leaseID    clientv3.LeaseID
	keepAliveC <-chan *clientv3.LeaseKeepAliveResponse
	serverInfo ServerInfo
}

func (s *ServerRegister) getKey() string {
	return etcdPrefix + "/" + s.serverInfo.ip + ":" + strconv.Itoa(s.serverInfo.port)
}

func (s *ServerRegister) getVal() string {
	return "http://" + s.serverInfo.ip + ":" + strconv.Itoa(s.serverInfo.port)
}

func NewServerRegister(endpoints []string, serverInfo ServerInfo, lease int64) (*ServerRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal("Conn Etcd error: ", err)
	}
	server := &ServerRegister{
		cli:        cli,
		serverInfo: serverInfo,
	}

	if err := server.PutKeyWithLease(lease); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *ServerRegister) PutKeyWithLease(lease int64) error {
	// 设置租约时间
	resp, err := s.cli.Grant(context.Background(), lease)
	if err != nil {
		return err
	}
	// 注册服务
	_, err = s.cli.Put(context.Background(), s.getKey(), s.getVal(), clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	// 设置续约，持续心跳监听
	leaseRespC, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	s.keepAliveC = leaseRespC
	log.Println("LeaseId: %d, Put key: %s, val: %s success!", s.leaseID, s.getKey(), s.getVal())
	return nil
}

func (s *ServerRegister) ListenRespC() {
	for leaseKeepAliveResp := range s.keepAliveC {
		log.Println("续约成功", leaseKeepAliveResp)
	}
	log.Println("停止续约")
}

func (s *ServerRegister) Close() error {
	if _, err := s.cli.Revoke(context.Background(), s.leaseID); err != nil {
		return err
	}
	log.Println("撤销租约")
	return s.cli.Close()
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
	ip, err := getLocalIP()
	if err != nil {
		log.Fatal(err)
		return
	}
	serverInfo := &ServerInfo{
		ip:   ip,
		port: 8888,
	}
	server, err := NewServerRegister(endpoints, *serverInfo, leaseTimeout)
	if err != nil {
		log.Fatal("New Server Register error: ", err)
	}

	// 响应监听
	go server.ListenRespC()
	select {
	// case <-time.After(20 * time.Second):
	// 	server.Close()
	}
}
