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
	weight     string
}

func NewServiceRegister(endpoints []string, addr, weight string, lease int64) (*ServiceRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	s := &ServiceRegister{
		cli:    cli,
		key:    "/" + schema + "/" + addr,
		weight: weight,
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
	_, err = s.cli.Put(context.Background(), s.key, s.weight, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	leaseRespC, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	s.keepAliveC = leaseRespC
	log.Println("Put key: %s weight: %s success!", s.key, s.weight)
	return nil
}

func (s *ServiceRegister) ListenLeaseRespC() {
	for leaseResp := range s.keepAliveC {
		log.Println("续约成功", leaseResp)
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
