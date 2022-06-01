package main

import (
	"context"
	"fmt"
	"log"

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

	res, err := client.Grant(context.TODO(), 5)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Put(context.TODO(), "/demo", "/test", clientv3.WithLease(res.ID))
	if err != nil {
		log.Fatal(err)
	}

	ch, kaerr := client.KeepAlive(context.TODO(), res.ID)
	if kaerr != nil {
		log.Fatal(kaerr)
	}
	for {
		ka := <-ch
		fmt.Println("ttl: ", ka.TTL)
	}
}
