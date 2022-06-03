package service

import (
	"context"
	"log"
	"sync"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"github.com/codetodo-io/storage-bootcamp/etcd/discovery/common"
)

type Master struct {
	cli        *clientv3.Client
	workerList map[string]string
	lock       sync.Mutex
}

func NewMaster(endpoints []string) (*Master, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: common.dialTimeout,
	})
	if err != nil {
		log.Fatal("Conn Etcd error: ", err)
		return nil, err
	}
	return &Master{
		cli:        cli,
		workerList: make(map[string]string),
	}, nil
}

func (s *Master) WatchServer(prefix string) error {
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, ev := range resp.Kvs {
		s.SetWorkerList(string(ev.Key), string(ev.Value))
	}

	go s.watcher(prefix)
	return nil
}

func (s *Master) watcher(prefix string) {
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Println("watching prefix: %s ...", prefix)
	for r := range rch {
		for _, ev := range r.Events {
			switch ev.Type {
			case mvccpb.PUT:
				s.SetWorkerList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				s.DelWorkerList(string(ev.Kv.Key))
			}
		}
	}
}

func (s *Master) SetWorkerList(key, val string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.workerList[key] = string(val)
	log.Println("Put key: %s, val: %s", key, val)
}

func (s *Master) DelWorkerList(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.workerList, key)
	log.Println("Del key: %s", key)
}

func (s *Master) GetWorkerList() []string {
	s.lock.Lock()
	defer s.lock.Unlock()
	addrs := make([]string, 0)

	for _, v := range s.workerList {
		addrs = append(addrs, v)
	}

	return addrs
}

func (s *Master) Close() error {
	return s.cli.Close()
}
