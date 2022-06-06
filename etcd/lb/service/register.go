package service

import (
	"context"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type ServiceRegister struct {
	cli        *clientv3.Client
	leaseID    clientv3.LeaseID
	keepAliveC <-chan *clientv3.LeaseKeepAliveResponse
	key        string
	val        string
}

func NewServiceRegister(endpoints []string, name, addr string, lease int64) (*ServiceRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	s := &ServiceRegister{
		cli: cli,
		key: "/" + schema + "/" + name + "/" + addr,
		val: addr,
	}
	if err := s.putKeyWithLease(lease); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *ServiceRegister) putKeyWithLease(lease int64) error {
	resp, err := s.cli.Grant(context.Background(), lease)
	if err != nil {
		return err
	}
	_, err = s.cli.Put(context.Background(), s.key, s.val, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	leaseRespC, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	s.keepAliveC = leaseRespC
	log.Println("Put key: %s val: %s success!", s.key, s.val)
	return nil
}

func (s *ServiceRegister) ListenLeaseRespC() {
	for leaseResp := range s.keepAliveC {
		log.Panicln("续约成功", leaseResp)
	}
	log.Println("关闭续约")
}

func (s *ServiceRegister) Close() error {
	if _, err := s.cli.Revoke(context.Background(), s.leaseID); err != nil {
		return err
	}
	log.Println("撤销续约")
	return s.cli.Close()
}
