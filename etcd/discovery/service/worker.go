package service

import (
	"context"
	"errors"
	"log"
	"net"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type WorkerInfo struct {
	IP   string
	Port int
}

type Worker struct {
	workerInfo WorkerInfo
	cli        *clientv3.Client
	leaseID    clientv3.LeaseID
	keepAliveC <-chan *clientv3.LeaseKeepAliveResponse
	prefix     string
}

func (s *Worker) getKey() string {
	return s.prefix + "/" + s.workerInfo.IP + ":" + strconv.Itoa(s.workerInfo.Port)
}

func (s *Worker) getVal() string {
	return "http://" + s.workerInfo.IP + ":" + strconv.Itoa(s.workerInfo.Port)
}

func NewWorker(endpoints []string, workerInfo WorkerInfo, dialTimeout time.Duration, prefix string, lease int64) (*Worker, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal("Conn Etcd error: ", err)
		return nil, err
	}
	worker := &Worker{
		cli:        cli,
		workerInfo: workerInfo,
		prefix:     prefix,
	}

	if err := worker.PutKeyWithLease(lease); err != nil {
		return nil, err
	}

	return worker, nil
}

func (s *Worker) PutKeyWithLease(lease int64) error {
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

func (s *Worker) ListenRespC() {
	for leaseKeepAliveResp := range s.keepAliveC {
		log.Println("续约成功", leaseKeepAliveResp)
	}
	log.Println("停止续约")
}

func (s *Worker) Close() error {
	if _, err := s.cli.Revoke(context.Background(), s.leaseID); err != nil {
		return err
	}
	log.Println("撤销租约")
	return s.cli.Close()
}

func GetLocalIP() (string, error) {
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
