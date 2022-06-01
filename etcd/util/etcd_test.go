package util

import (
	"etcd/util/etcd"
	"fmt"
	"time"
)

func main() {
	cli, err := etcd.CreateCli([]string{"etcd:1279"}, 2)
	defer etcd.Close()

	if err != nil {
		fmt.Println(err)
	}

	ptr := cli.PutValue(cli, "abc", "123")
	fmt.Println("推送成功", ptr)
	gtr := cli.GetValue(cli, "abc")
	fmt.Println("查询结果，abc=%s", gtr)
	cli.DelValue(cli, "abc")
	gtr = cli.GetValue(cli, "abc")
	fmt.Println("查询结果，abc=%s", gtr)

	ptr = cli.PutValue(cli, "abcdef", "123456")
	gtr_pre := cli.GetValueByPrefix(cli, "abc")
	fmt.Println("查询结果，abc=%s", gtr_pre)
	dtr_pre := cli.DelValueByPrefix(cli, "abc")
	fmt.Println("删除结果，%s", dtr_pre)
	gtr = cli.GetValue(cli, "abcdef")
	fmt.Println("abcdef=%s", gtr)

	ptr = cli.PutValue(cli, "abc", "123")
	deal := cli.OfficeDeal(cli, "abc", "123", "456")
	fmt.Println("事务结果，%s", deal)
	gtr = cli.GetValue(cli, "abc")
	fmt.Println("查询结果，abc=%s", gtr)
	deal = cli.OfficeDeal(cli, "abc", "456", "789")
	fmt.Println("事务结果，%s", deal)
	gtr = cli.GetValue(cli, "abc")
	fmt.Println("查询结果，abc=%s", gtr)

	var c1 chan bool
	var c2 chan bool

	go func() {
		m := cli.WatchKey(cli, "abc")
		fmt.Println("监控键值，%s", m)
		c1 <- true
	}()

	go func() {
		m := cli.WatchKeyByPrefix(cli, "a")
		fmt.Println("带前缀监控键值，%s", m)
		c2 <- true
	}()

	cli.PutValue(cli, "abc", "12345")
	cli.PutValue(cli, "abc", "54321")

	cycle := cli.CycleSetKeyValueByTime(cli, 10, "haha", "pengpeng")
	fmt.Println("周期结果，%s", cycle)
	gtr = cli.GetValue(cli, "haha")
	fmt.Println("周期结束之前", gtr)
	time.Sleep(11 * time.Second)
	gtr = cli.GetValue(cli, "haha")
	fmt.Println("周期结束之后", gtr)

	<-c1
	<-c2
}
