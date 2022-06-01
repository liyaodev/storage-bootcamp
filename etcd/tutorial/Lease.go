package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
)

var (
	dialTimeout    = 3 * time.Second
	requestTimeout = 10 * time.Second
)

func LeaseDemo(ctx context.Context, cli *clientv3.Client, kv clientv3.KV) {
	kv.Delete(ctx, "demo", clientv3.WithPrefix())

	g, _ := kv.Get(ctx, "demo")
	if len(g.Kvs) == 0 {
		fmt.Println("The `demo` is empty.")
	}

	res, err := cli.Grant(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}

	// Insert key with a lease of 5 second TTL
	kv.Put(ctx, "demo", "hello", clientv3.WithLease(res.ID))
	g, _ = kv.Get(ctx, "demo")
	if len(g.Kvs) == 1 {
		fmt.Println("found `demo`, ", g)
	}
	// Let the TTL expire
	time.Sleep(5 * time.Second)
	g, _ = kv.Get(ctx, "demo")
	if len(g.Kvs) == 0 {
		fmt.Println("The `demo` is empty.")
	}
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"etcd:2379"},
		DialTimeout: dialTimeout,
	})
	if err != nil {
		fmt.Println("connect to etcd failed, err: %v\n", err)
		return
	}
	fmt.Println("connect to etcd success")
	defer cli.Close()
	kv := clientv3.NewKV(cli)

	fmt.Println("*** Lease Demo Start ***")
	LeaseDemo(ctx, cli, kv)
	fmt.Println("*** Lease Demo End ***")
}
