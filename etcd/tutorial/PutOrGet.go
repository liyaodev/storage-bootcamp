package main

import (
	"context"
	"fmt"

	"go.etcd.io/etcd/clientv3"
)

func main() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"etcd:2379"},
		DialTimeout: 5 * time.Sencond,
	})

	if err != nil {
		fmt.Printf("connect to etcd failed, err: %v\n", err)
		return
	}

	fmt.Printf("connect to etcd success")
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Sencond)
	_, err = client.Put(ctx, "demo", "test")
	cancel()
	if err != nil {
		fmt.Printf("put to etcd failed, err: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Sencond)
	res, err := client.Get(ctx, "demo")
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed, err: %v\n", err)
		return
	}

	for _, e := range res.Kvs {
		fmt.Printf("%s:%s\n", e.Key, e.Value)
	}
}
