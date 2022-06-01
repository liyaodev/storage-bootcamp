package util

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

func CreateCli(endPoint []string, dialTimeout int) (*clientv3.Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endPoint,
		DialTimeout: time.Duration(dialTimeout) * time.Second,
	})

	if err != nil {
		return nil, err
	}
	return cli, err
}

func Close(cli *clientv3.Client) {
	cli.Close()
}

// 插入
func PutValue(cli *clientv3.Client, key, value string) bool {
	_, err := cli.Put(context.TODO(), key, value)
	if err != nil {
		return false
	}
	return true
}

// 查询
func GetValue(cli *clientv3.Client, key string) []*mvccpb.KeyValue {
	res, err := cli.Get(context.TODO(), key)
	if err != nil {
		fmt.Println(err)
	} else {
		return res.Kvs
	}
	return nil
}

// 按前缀批量查询
func GetValueByPrefix(cli *clientv3.Client, key string) []*mvccpb.KeyValue {
	res, err := cli.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		fmt.Println(err)
	} else {
		return res.Kvs
	}
	return nil
}

// 删除
func DelValue(cli *clientv3.Client, key string) int64 {
	res, err := cli.Delete(context.TODO(), key)
	if err != nil {
		fmt.Println(err)
	} else {
		return res.Deleted
	}
	return 0
}

// 按前缀批量删除
func DelValueByPrefix(cli *clientv3.Client, key string) int64 {
	res, err := cli.Delete(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		fmt.Println(err)
	} else {
		return res.Deleted
	}
	return 0
}

// 事务
func OfficeDeal(cli *clientv3.Client, key, value, otherValue string) bool {
	_, err := cli.Txn(context.TODO()).
		If(clientv3.Compare(clientv3.Value(key), "=", value)).
		Then(clientv3.OpPut(key, value)).
		Else(clientv3.OpPut(key, otherValue)).Commit()
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// 监听
func WatchKey(cli *clientv3.Client, key string) []byte {
	rch := cli.WatchKey(context.Background(), key)
	for res := range rch {
		for _, ev := range res.Events {
			fmt.Println("watch: %s: %q: %q", ev.Type, ev.Kv.Key, ev.Kv.Value)
			return ev.Kv.Value
		}
	}
	return nil
}

// 按前缀监听
func WatchKeyByPrefix(cli *clientv3.Client, key string) []byte {
	rch := cli.WatchKey(context.Background(), key, clientv3.WithPrefix())
	for res := range rch {
		for _, ev := range res.Events {
			fmt.Println("watch: %s: %q: %q", ev.Type, ev.Kv.Key, ev.Kv.Value)
			return ev.Kv.Value
		}
	}
	return nil
}

// 按范围监听
func WatchKeyByRange(cli *clientv3.Client, startKey, endKey string) []byte {
	rch := cli.WatchKey(context.Background(), startKey, clientv3.WithRange(endKey))
	for res := range rch {
		for _, ev := range res.Events {
			fmt.Println("watch: %s: %q: %q", ev.Type, ev.Kv.Key, ev.Kv.Value)
			return ev.Kv.Value
		}
	}
	return nil
}

// 周期
func CycleSetKeyValueByTime(cli *clientv3.Client, timeOut int64, key, value, string) bool {
	res, err := cli.Grant(context.TODO(), timeOut)
	if err != nil {
		fmt.Println(err)
		return false
	}

	_, err = cli.Put(context.TODO(), key, value, clientv3.WithLease(res.ID))
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}