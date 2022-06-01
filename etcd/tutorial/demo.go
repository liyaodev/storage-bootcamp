package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
)

const (
	prefix = "/demo/"
)

type EtcdConfig struct {
	Endpoints []string
	KeyFile   string
	CertFile  string
	CAFile    string
	Username  string
	Password  string
}

func newEtcdClient(c *EtcdConfig) (*etcd.Client, error) {
	var tlsInfo *tls.Config
	var pool *x509.CertPool

	if c.CertFile != "" && c.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("Unable to load cert: %v", err)
		}

		if c.CAFile != "" {
			caData, err := ioutil.ReadFile(c.CAFile)
			if err != nil {
				return nil, fmt.Errorf("Unable to load ca file context: %v", caData)
			}
			pool = x509.NewCertPool()
			pool.AppendCertsFromPEM(caData)
		}

		tlsInfo = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      pool,
		}
	}

	client, err := etcd.New(etcd.Config{
		Endpoints: c.Endpoints,
		Username:  c.Username,
		Password:  c.Password,
		TLS:       tlsInfo,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

var client *etcd.Client

func init() {
	var err error
	client, err = newEtcdClient(&EtcdConfig{
		Endpoints: []string{"http://etcd:2379"},
	})

	if err != nil {
		panic(fmt.Sprintf("Unable to create etcd client: %v", err))
	}
}

func setWinner(who, what string) error {
	key := prefix + who
	fmt.Println("Set %s to [%s]", key, what)

	_, err := client.Put(context.Background(), key, what)
	if err != nil {
		fmt.Println("Put error: %v", err)
	}
	return err
}

func setWinnerWithLease(who, what string, ttl int) error {
	key := prefix + who
	fmt.Println("Set %s to [%s] with lease %d", key, what, ttl)

	res, err := client.Grant(context.Background(), int64(ttl))
	if err != nil {
		fmt.Println("Unable to create lease")
	}
	_, err = client.Put(context.Background(), key, what, etcd.WithLease(res.ID))
	if err != nil {
		fmt.Println("Put error: %v", err)
	}
	return err
}

func deleteWinner(who string) error {
	key := prefix + who
	fmt.Println("Delete %s", key)

	_, err := client.Delete(context.Background(), key)
	if err != nil {
		fmt.Printf("Delete error: %v", err)
	}
	return err
}

func getWinner(who string) error {
	key := prefix + who
	res, err := client.Get(context.Background(), key)
	if err != nil {
		return err
	}

	if len(res.Kvs) == 0 {
		fmt.Println("%s does not win Awards", who)
		return nil
	}

	tokens := strings.Split(string(res.Kvs[0].Key), "/")
	who = tokens[len(tokens)-1]
	what := res.Kvs[0].Value
	fmt.Println("%s won awards for %s", who, what)
	return nil
}

func getWinners() error {
	key := prefix
	res, err := client.Get(context.Background(), key, etcd.WithPrefix())
	if err != nil {
		return err
	}
	for _, ev := range res.Kvs {
		fmt.Println("%s won awards for %s", ev.Key, ev.Value)
	}
	return nil
}

func updateWinner(who string, prev, new string) error {
	key := prefix + who
	fmt.Println("Update %s to [%s]", key, new)

	res, err := client.Txn(context.Background()).
		If(etcd.Compare(etcd.Value(key), "=", prev)).
		Then(etcd.OpPut(key, new)).
		Commit()
	if err != nil {
		fmt.Println("Set value error: %v", err)
	}

	if !res.Succeeded {
		fmt.Println("Set value failed: value compare error")
	}
	return err
}

func watchWinners(closeCh <-chan struct{}) {
	key := prefix
	resultCh := client.Watch(context.Background(), key, etcd.WithPrefix())

	for {
		select {
		case res := <-resultCh:
			for _, event := range res.Events {
				fmt.Println("Event received: %s %q: %q", event.Type, event.Kv.Key, event.Kv.Value)
			}
		case <-closeCh:
			return
		}
	}
}

type winner struct {
	name string
	what string
}

type demo struct {
	action      func() error
	description string
}

func demoGetSet() error {
	who := "demo-test"
	err := setWinner(who, "Artificial Interlligence")
	if err != nil {
		return err
	}

	err = getWinner(who)
	if err != nil {
		return err
	}

	err = deleteWinner(who)
	if err != nil {
		return err
	}

	err = getWinner(who)
	if err != nil {
		return err
	}

	return nil
}

func demoPrefix() error {
	winners := []winner{
		winner{
			"Dijkstra",
			"Programming Languages",
		},
		winner{
			"Knuth",
			"analysis of algorithms",
		},
	}
	for _, w := range winners {
		err := setWinner(w.name, w.what)
		if err != nil {
			return err
		}
	}

	return getWinners()
}

func demoTransaction() error {
	who := "demo-Lee"
	err := setWinner(who, "WWW")
	if err != nil {
		return err
	}

	_ = updateWinner(who, "lisp", "Inventing World Wide Web")
	getWinner(who)

	err = updateWinner(who, "WWW", "Inventing World Wide Web")
	if err != nil {
		return err
	}

	return getWinner(who)
}

func demoWatch() error {
	closeC := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		watchWinners(closeC)
		wg.Done()
	}()

	winners := []winner{
		winner{
			"Ken-Thompson",
			"Unix operating system",
		},
		winner{
			"Alan-Key",
			"Object-Oriented Programming Languages",
		},
		winner{
			"cizixs",
			"nothing",
		},
	}

	time.Sleep(time.Microsecond * 500)
	for _, w := range winners {
		err := setWinner(w.name, w.what)
		if err != nil {
			return err
		}
	}

	deleteWinner("cizixs")

	closeC <- struct{}{}
	wg.Wait()

	return nil
}

func demoLease() error {
	err := setWinnerWithLease("cizixs", "nothing", 5)
	if err != nil {
		return err
	}

	if err := getWinner("cizixs"); err != nil {
		return err
	}

	fmt.Println("wait for lease to expire...")
	time.Sleep(time.Second * 7)

	return getWinner("cizixs")
}

func main() {
	defer client.Close()
	demos := []demo{
		demo{
			action:      demoGetSet,
			description: "simple set and get",
		},
		demo{
			action:      demoPrefix,
			description: "use prefix to get all values",
		},
		demo{
			action:      demoTransaction,
			description: "multiple action with transaction",
		},
		demo{
			action:      demoWatch,
			description: "watch key changes",
		},
		demo{
			action:      demoLease,
			description: "assign lease to keys",
		},
	}

	for i, d := range demos {
		fmt.Println("********** DEMO %d: %s **********", i+1, d.description)
		err := d.action()
		if err != nil {
			fmt.Println("Demo failed: %v", err)
		}
	}
}
