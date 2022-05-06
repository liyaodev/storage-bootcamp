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

	w := client.Watch(context.Background(), "demo")
	for res := range w {
		for _, e := range res.Events {
			fmt.Printf("Type: %s, Key: %s, Value: %s\n", e.Type, e.Kv.Key, e.Kv.Value)
		}
	}
}
