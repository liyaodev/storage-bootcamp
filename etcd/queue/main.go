package main

import (
	"fmt"
	"log"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	recipe "go.etcd.io/etcd/client/v3/experimental/recipes"
)

func main() {
	endpoints := os.Getenv("ETCD_ENDPOINTS")
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{endpoints},
	})
	if err != nil {
		log.Fatalf("error New (%v)", err)
	}

	go func() {
		queue := recipe.NewQueue(client, "queue-demo")
		for i := 0; i < 5; i++ {
			if err := queue.Enqueue(fmt.Sprintf("%d", i)); err != nil {
				log.Fatalf("Enqueue error (%v)", err)
			}
		}
	}()

	go func() {
		queue := recipe.NewQueue(client, "queue-demo")
		for i := 10; i < 100; i++ {
			if err := queue.Enqueue(fmt.Sprintf("%d", i)); err != nil {
				log.Fatalf("Enqueue error (%v)", err)
			}
		}
	}()

	queue := recipe.NewQueue(client, "queue-demo")
	for i := 0; i < 100; i++ {
		str, err := queue.Dequeue()
		if err != nil {
			log.Fatalf("Dequeue error (%v)", err)
		}
		fmt.Println(s)
	}

	time.sleep(time.Second * 3)
}
