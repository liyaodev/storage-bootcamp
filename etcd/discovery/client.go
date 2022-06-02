package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

var (
	dialTimeout  = 5 * time.Second
	leaseTimeout = int64(3)
	etcdPrefix   = "/etcd/discovery"
)

type ServerDiscovery struct {
	cli        *clientv3.Client
	serverList map[string]string
	lock       sync.Mutex
}

func NewServerDiscovery(endpoints []string) *ServerDiscovery {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal("Conn Etcd error: ", err)
	}
	return &ServerDiscovery{
		cli:        cli,
		serverList: make(map[string]string),
	}
}

func (s *ServerDiscovery) WatchServer(prefix string) error {
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, ev := range resp.Kvs {
		s.SetServerList(string(ev.Key), string(ev.Value))
	}

	go s.watcher(prefix)
	return nil
}

func (s *ServerDiscovery) watcher(prefix string) {
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Println("watching prefix: %s ...", prefix)
	for r := range rch {
		for _, ev := range r.Events {
			switch ev.Type {
			case mvccpb.PUT:
				s.SetServerList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				s.DelServerList(string(ev.Kv.Key))
			}
		}
	}
}

func (s *ServerDiscovery) SetServerList(key, val string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.serverList[key] = string(val)
	log.Println("Put key: %s, val: %s", key, val)
}

func (s *ServerDiscovery) DelServerList(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.serverList, key)
	log.Println("Del key: %s", key)
}

func (s *ServerDiscovery) GetServerList() []string {
	s.lock.Lock()
	defer s.lock.Unlock()
	addrs := make([]string, 0)

	for _, v := range s.serverList {
		addrs = append(addrs, v)
	}

	return addrs
}

func (s *ServerDiscovery) Close() error {
	return s.cli.Close()
}

func main() {
	var endpoints = []string{"etcd:2379"}
	server := NewServerDiscovery(endpoints)
	defer server.Close()

	server.WatchServer(etcdPrefix)

	for {
		select {
		case <-time.Tick(10 * time.Second):
			log.Println("serverList: ", server.GetServerList())
		}
	}
}
