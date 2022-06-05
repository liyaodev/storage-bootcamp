package main

import (
	"fmt"
	"time"

	"github.com/codetodo-io/storage-bootcamp/etcd/shareconfig/service"
)

var (
	dialTimeout = 5 * time.Second
	etcdPrefix  = "/etcd/shareconfig"
)

func main() {
	c, err := service.New([]string{"etcd:2379"}, dialTimeout, etcdPrefix)
	if err != nil {
		println(err)
	}
	fmt.Println("Config bar is", c.Get("bar"))
}
