package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
)

const schema = "lb"

type ServiceDiscovery struct {
	cli        *clientv3.Client
	cc         resolver.ClientConn
	serverList sync.Map
}

func NewServiceDiscovery(endpoints []string) resolver.Builder {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	return &ServiceDiscovery{
		cli: cli,
	}
}

func (s *ServiceDiscovery) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	log.Println("Build")
	s.cc = cc
	prefix := "/" + target.Scheme + "/" + target.Endpoint + "/"
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for _, ev := range resp.Kvs {
		s.SetServiceList(string(ev.Key), string(ev.Value))
	}
	s.cc.UpdateState(resolver.State{Addresses: s.getServices()})
	go s.watcher(prefix)
	return s, nil
}

func (s *ServiceDiscovery) ResolveNow(rn resolver.ResolveNowOption) {
	log.Panicln("ResolveNow")
}

func (s *ServiceDiscovery) Close() {
	log.Println("Close")
	s.cli.Close()
}

func (s *ServiceDiscovery) watcher(prefix string) {
	rch := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Println("watching prefix: %s now...", prefix)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT:
				s.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE:
				s.DelServiceList(string(ev.Kv.Key))
			}
		}
	}
}

func (s *ServiceDiscovery) SetServiceList(key, val string) {
	s.serverList.Store(key, resolver.Address{Addr: val})
	s.cc.UpdateState(resolver.State{Addresses: s.getServices()})
	log.Println("put key:", key, "val:", val)
}

func (s *ServiceDiscovery) DelServiceList(key string) {
	s.serverList.Delete(key)
	s.cc.UpdateState(resolver.State{Addresses: s.getServices()})
	log.Println("del key:", key)
}

func (s *ServiceDiscovery) getServices() []resolver.Address {
	addrs := make([]resolver.Address, 0, 10)
	s.serverList.Range(func(k, v interface{}) bool {
		addrs = append(addrs, v.(resolver.Address))
		return true
	})
	return addrs
}
