package service

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
)

type Config struct {
	kv     map[string]string
	cli    *clientv3.Client
	rch    clientv3.WatchChan
	lock   sync.RWMutex
	prefix string
}

func New(endpoints []string, dialTimeout time.Duration, prefix string) (*Config, error) {
	var err error
	c := &Config{
		kv:     map[string]string{},
		prefix: prefix,
	}

	c.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal("Conn Etcd error: ", err)
		return nil, err
	}

	err = c.initAndWatch()
	if err != nil {
		c.cli.Close()
		return nil, err
	}

	go func() {
		for {
			for wresp := range c.rch {
				for _, ev := range wresp.Events {
					c.set(ev.Kv.Key, ev.Kv.Value)
				}
			}
			log.Println("Config watch channel closed")
			for {
				err = c.initAndWatch()
				if err == nil {
					break
				}
				log.Println("Config get failed: ", err)
				time.Sleep(time.Second)
			}
		}
	}()

	return c, nil
}

func (c *Config) initAndWatch() error {
	c.rch = c.cli.Watch(context.TODO(), c.prefix, clientv3.WithPrefix())
	resp, err := c.cli.Get(context.TODO(), c.prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, kv := range resp.Kvs {
		c.set(kv.Key, kv.Value)
	}
	return nil
}

func (c *Config) set(key, value []byte) {
	strKey := strings.TrimPrefix(string(key), c.prefix)
	c.lock.Lock()
	defer c.lock.Unlock()
	if len(value) == 0 {
		delete(c.kv, string(strKey))
	} else {
		c.kv[string(strKey)] = string(value)
	}
}

func (c *Config) Get(key string) string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.kv[key]
}

func (c *Config) String() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	b, _ := json.MarshalIndent(c.kv, "", " ")
	return string(b)
}
