package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var (
	dialTimeout    = 3 * time.Second
	requestTimeout = 10 * time.Second
)

func WatchDemo(ctx context.Context, cli *clientv3.Client, kv clientv3.KV) {
	// Delete all keys
	kv.Delete(ctx, "demo", clientv3.WithPrefix())

	stopC := make(chan interface{})
	go func() {
		watchC := cli.Watch(ctx, "demo", clientv3.WithPrefix())
		for true {
			select {
			case result := <-watchC:
				for _, e := range result.Events {
					fmt.Println("%s %q: %q\n", e.Type, e.Kv.Key, e.Kv.Value)
				}
			case <-stopC:
				fmt.Println("Done watching.")
				return
			}
		}
	}()

	// Inset Keys
	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("demo_%02d", i)
		kv.Put(ctx, k, strconv.Itoa(i))
	}

	// Make sure watcher go routine has time to recive PUT events
	time.Sleep(time.Second)
	stopC <- 1
	// Insert some more keys (no one is watching)
	for i := 10; i < 20; i++ {
		k := fmt.Sprintf("demo_%02d", i)
		kv.Put(ctx, k, strconv.Itoa(i))
	}
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	cli, err := clientv3.New(clientv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   []string{"etcd:2379"},
	})
	if err != nil {
		fmt.Println("connect to etcd failed, err: %v\n", err)
		return
	}
	fmt.Println("connect to etcd success")
	defer cli.Close()
	kv := clientv3.NewKV(cli)

	fmt.Println("*** Watch Demo Start ***")
	WatchDemo(ctx, cli, kv)
	fmt.Println("*** Watch Demo End ***")

}
