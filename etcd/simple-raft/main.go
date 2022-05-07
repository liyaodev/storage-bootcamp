package main

import (
	"flag"
	"strings"
)

func main() {
	port := flag.String("port", ":8001", "rpc listen port")
	cluster := flag.String("cluster", "127.0.0.1:8001", "rpc listen uri")
	id := flag.Int("id", 1, "Raft Node ID")
	flag.Parse()

	clusters := strings.Split(*cluster, ",")
	ns := make(map[int]*Node)
	for k, v := range clusters {
		ns[k] = newNode(v)
	}

	// 创建节点
	raft := &RaftNode{}
	raft.cid = *id
	raft.otherNodes = ns

	// 监听 RPC
	raft.rpc(*port)
	raft.start()

	select {}
}
