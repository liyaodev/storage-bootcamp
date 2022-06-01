package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"go.etcd.io/etcd/clientv3"
)

var (
	dialTimeout    = 3 * time.Second
	requestTimeout = 10 * time.Second
)

func PutOrGetDemo(ctx context.Context, kv clientv3.KV) {
	kv.Delete(ctx, "demo", clientv3.WithPrefix())
	// Insert a key value
	p, err := kv.Put(ctx, "demo", "hello")
	if err != nil {
		log.Fatal(err)
	}
	rev := p.Header.Revision
	fmt.Println("Revision:", rev)
	g, err := kv.Get(ctx, "demo")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Value: ", string(g.Kvs[0].Value), "Revision: ", g.Header.Revision)
	// Modify the value of an existing key (create new revision)
	kv.Put(ctx, "demo", "hello world!")
	g, _ = kv.Get(ctx, "demo")
	fmt.Println("Value: ", string(g.Kvs[0].Value), "Revision: ", g.Header.Revision)
	// Get the value of the previous revision
	g, _ = kv.Get(ctx, "demo", clientv3.WithRev(rev))
	fmt.Println("Value: ", string(g.Kvs[0].Value), "Revision: ", g.Header.Revision)
}

func MultiPutOrGetDemo(ctx context.Context, kv clientv3.KV) {
	kv.Delete(ctx, "demo", clientv3.WithPrefix())
	// Insert 20 keys
	for i := 0; i < 20; i++ {
		k := fmt.Sprintf("demo_%02d", i)
		kv.Put(ctx, k, strconv.Itoa(i))
	}
	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		clientv3.WithLimit(10),
	}
	g, err := kv.Get(ctx, "demo", opts...)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("--- First Page ---")
	for _, item := range g.Kvs {
		fmt.Println(string(item.Key), string(item.Value))
	}
	lastKey := string(g.Kvs[len(g.Kvs)-1].Key)
	fmt.Println("--- Second Page ---")
	opts = append(opts, clientv3.WithFromKey())
	g, _ = kv.Get(ctx, lastKey, opts...)
	// Skipping the first item, which the last item from the previous Get
	for _, item := range g.Kvs[1:] {
		fmt.Println(string(item.Key), string(item.Value))
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

	fmt.Println("*** PutOrGetDemo Start ***")
	PutOrGetDemo(ctx, kv)
	fmt.Println("*** PutOrGetDemo End ***")

	fmt.Println("*** MultiPutOrGetDemo Start ***")
	MultiPutOrGetDemo(ctx, kv)
	fmt.Println("*** MultiPutOrGetDemo End ***")
}
