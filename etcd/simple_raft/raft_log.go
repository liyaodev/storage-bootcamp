package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/rpc"
	"time"
)

// 节点结构体定义
type Node struct {
	connect bool
	address string
}

// 节点状态定义
type RaftState int

// Raft 三种状态：Follower、Candidate、Leader
const (
	Follower RaftState = iota + 1
	Candidate
	Leader
)

// LogEntry
type LogEntry struct {
	LogTerm  int
	LogIndex int
	LogCMD   interface{}
}

// RaftNode 结构体定义
type RaftNode struct {
	// 节点 ID
	cid int
	// 其他节点信息
	otherNodes map[int]*Node
	// 当前节点状态
	currentState RaftState
	// 当前任期
	currentTerm int
	// 当前任期投票给了谁，初始为 -1
	votedFor int
	// 当前任期获得的投票数量
	voteCnt int
	// heartbeat channel
	heartbeatC chan bool
	// to leader channel
	toLeaderC chan bool

	// 日志条目集合
	log []LogEntry
	// 被提交的最大索引
	commitIndex int
	// 被应用到状态机的最大索引
	lastApplied int

	// 保存需要发送给每个节点的下一个条目索引
	nextIndex []int
	// 保存已经复制给每个节点日志的最高索引
	matchIndex []int
}

// 新建节点
func newNode(address string) *Node {
	node := &Node{}
	node.address = address
	return node
}

// Request
type VoteArgs struct {
	// 当前任期号
	CurrentTerm int
	// 候选人 ID
	CandidateID int
}

// Response
type VoteReply struct {
	// 当前任期号，如果其他节点任期比候选者大，Candidate以此为更新
	CurrentTerm int
	// 候选人投票状态
	VoteGranted bool
}

// 投票广播
func (rf *RaftNode) broadcastRequestVote() {
	// 设置 request
	var args = VoteArgs{
		CurrentTerm: rf.currentTerm,
		CandidateID: rf.cid,
	}
	// 通过 rf.otherNodes 遍历广播投票申请
	for i := range rf.otherNodes {
		go func(i int) {
			var reply VoteReply
			rf.sendRequestVote(i, args, &reply)
		}(i)
	}
}

// 发送请求
func (rf *RaftNode) sendRequestVote(cid int, args VoteArgs, reply *VoteReply) {
	// 创建 client
	client, err := rpc.DialHTTP("tcp", rf.otherNodes[cid].address)
	if err != nil {
		log.Fatal("create client error: ", err)
	}
	defer client.Close()
	// 调用 Follower 节点的 RequestVote 方法
	client.Call("RaftNode.RequestVote", args, reply)
	// 如果 Candiate 任期小于 Follower，当前选举无效 Candidate 转换为 Follower
	if reply.CurrentTerm > rf.currentTerm {
		rf.currentTerm = reply.CurrentTerm
		rf.currentState = Follower
		rf.votedFor = -1
		return
	}
	// 成功获选
	if reply.VoteGranted {
		// 票数 += 1
		rf.voteCnt += 1
		// 获取票数大于集群一半即获选(len(rf.nodes) / 2 + 1)
		if rf.voteCnt > len(rf.otherNodes)/2+1 {
			rf.toLeaderC <- true
		}
	}
}

// Follower 处理投票申请
func (rf *RaftNode) RequestVote(args VoteArgs, reply *VoteReply) error {
	// 如果 Candidate 的 term 小于 Follower，拒绝投票
	if args.CurrentTerm < rf.currentTerm {
		reply.CurrentTerm = rf.currentTerm
		reply.VoteGranted = false
		return nil
	}
	// Follower 未投票则投给 Candidiate
	if rf.votedFor == -1 {
		rf.currentTerm = args.CurrentTerm
		rf.votedFor = args.CandidateID
		reply.CurrentTerm = rf.currentTerm
		reply.VoteGranted = true
		return nil
	}
	// 其他情况
	reply.CurrentTerm = rf.currentTerm
	reply.VoteGranted = false
	return nil
}

type HeartbeatArgs struct {
	// 当前 Leader Term
	Term int
	// 当前 Leader id
	LeaderID int

	PrevLogIndex int
	PrevLogTerm  int
	// 待存储的日志条目，纯心跳请求时为空
	Entries []LogEntry
	// Leader 已 commit 的索引值
	LeaderCommit int
}

type HeartbeatReply struct {
	// 当前 Follower Term
	Term int
	// 是否心跳成功
	Success bool
	// 如果 Follower Index 小于 Leader Index，告诉 Leader 下次发送的索引位置
	NextIndex int
}

// 广播心跳
func (rf *RaftNode) broadcastHeartbeat() {
	// 遍历所有节点
	for i := range rf.otherNodes {
		args := HeartbeatArgs{
			Term:     rf.currentTerm,
			LeaderID: rf.cid,
			LeaderCommit = rf.commitIndex
		}

		// 计算 prevLogIndex、prevLogTerm
		prevLogIndex := rf.nextIndex[i] - 1
		// 如果没有可以发送的 LogEntry
		if rf.getLastIndex() > prevLogIndex {
			// 提取 prevLogIndex、prevLogTerm 之后的 Entry，发送给 Follower
			args.PrevLogIndex = prevLogIndex
			args.PrevLogTerm = rf.log[prevLogIndex].LogTerm
			args.Entries = rf.log[prevLogIndex:]
			log.Println("send entries: %v", args.Entries)
		}
		// 发送给某一个节点
		go func(i int, args HeartbeatArgs) {
			var reply HeartbeatReply
			rf.sendHeartbeat(i, args, &reply)
		}(i, args)
	}
}

// 发送心跳请求
func (rf *RaftNode) sendHeartbeat(cid int, args HeartbeatArgs, reply *HeartbeatReply) {
	// 创建 client
	client, err := rpc.DialHTTP("tcp", rf.otherNodes[cid].address)
	if err != nil {
		log.Fatal("create client error: ", err)
	}
	defer client.Close()
	// 调用 Follower 节点的 Heartbeat 方法
	client.Call("RaftNode.Heartbeat", args, reply)
	// 如果 Follower 任期大于 Leader，代表 Leader 过时转换为 Follower
	// if reply.Term > rf.currentTerm {
	// 	rf.currentTerm = reply.Term
	// 	rf.currentState = Follower
	// 	rf.votedFor = -1
	// }
	// 如果 Leader 节点落后于 Follower 节点
	if reply.Success {
		// 如果 heartbeat 中带有 LogEntry 实体，更新对应节点的 nextIndex 和 matchIndex
		if reply.NextIndex > 0 {
			rf.nextIndex[cid] = reply.NextIndex
			rf.matchIndex[cid] = reply.nextIndex[cid] - 1
			// TODO
			// 如果大于半数节点同步成功
			// 1. 更新 Leader commitIndex
			// 2. 返回客户端
			// 3. 应用状态机
			// 4. 通知 Followers Entry 操作
		}
	} else {
		// 如果 Follower 任期大于 Leader，代表 Leader 过时转换为 Follower
		if reply.Term > rf.currentTerm {
			rf.currentTerm = reply.Term
			rf.currentState = Follower
			rf.votedFor = -1
			return 
		}
	}
}

func (rf *RaftNode) getLastIndex() int {
	rlen := len(rf.log)
	if rlen == 0 {
		return 0
	}
	return rf.log[rlen - 1].LogIndex
}

func (rf *RaftNode) getLastTerm() int {
	rlen := len(rf.log)
	if rlen == 0 {
		return 0
	}
	return rf.log[rlen - 1].LogTerm
}

func (rf *RaftNode) Heartbeat(args HeartbeatArgs, reply *HeartbeatReply) error {
	// 如果 Leader 节点 Term 小于 Follower，不做处理并返回
	if args.Term < rf.currentTerm {
		reply.Term = rf.currentTerm
		return nil
	}
	// 如果 Leader 节点 Term 大于 Follower，说明 Follower 过时，直接重置
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.currentState = Follower
		rf.votedFor = -1
	}
	// 返回当前节点 Follower 给 Leader
	reply.Term = rf.currentTerm
	// 心跳成功，发送消息给 heartbeatC
	rf.heartbeatC <- true
	// 如果没有 entries，只有 heartbeat
	if len(args.Entries) == 0 {
		reply.Success = true
		reply.Term = rf.currentTerm
		return nil
	}
	// 如果有 entries
	// Leader 维护的 LogIndex 大于当前 Follower 的 LogIndex，代表 Follower 有延迟，需要告知 Leader 当前最大索引
	if args.PrevLogIndex > rf.getLastIndex() {
		reply.Success = false
		reply.Term = rf.currentTerm
		reply.NextIndex = rf.getLastIndex() + 1
		return nil
	}
	rf.log = append(rf.log, args.Entries...)
	rf.commitIndex = rf.getLastIndex()
	reply.Success = true
	reply.Term = rf.currentTerm
	reply.NextIndex = rf.getLastIndex() + 1

	return nil
}

func (rf *RaftNode) rpc(port string) {
	rpc.Register(rf)
	rpc.HandleHTTP()
	go func() {
		err := http.ListenAndServe(port, nil)
		if err != nil {
			log.Fatal("lister error: ", err)
		}
	}()
}

// 创建 RaftNode
func (rf *RaftNode) start() {
	// 初始化 Raft
	rf.currentState = Follower
	rf.currentTerm = 0
	rf.votedFor = -1
	rf.heartbeatC = make(chan bool)
	rf.toLeaderC = make(chan bool)
	// 节点状态变更以及 RPC 处理
	go func() {
		rand.Seed(time.Now().UnixNano())
		// 持续处理节点任务和通信
		for {
			switch rf.currentState {

			case Follower:
				fmt.Println("This is Follower[%d]", rf.cid)
				select {
				// 接收心跳
				case <-rf.heartbeatC:
					log.Println("Follower[%d] recived heartbeat.", rf.cid)
				// 心跳超时，状态转换为 Candidate
				case <-time.After(time.Duration(rand.Intn(500-300)+300) * time.Microsecond):
					log.Println("Follower[%d] timeout", rf.cid)
					rf.currentState = Candidate
				}
			case Candidate:
				fmt.Println("This is Candidate[%d]", rf.cid)
				// 1. term += 1
				rf.currentTerm += 1
				// 2. 为自己投票
				rf.votedFor = rf.cid
				rf.voteCnt = 1
				// 3. 广播拉票
				go rf.broadcastRequestVote()
				select {
				// 选举超时，状态转换为 Follower
				case <-time.After(time.Duration(rand.Intn(5000-300)+300) * time.Microsecond):
					rf.currentState = Follower
				// 选举成功
				case <-rf.toLeaderC:
					fmt.Println("This is new Leader[%d]", rf.cid)
					rf.currentState = Leader

					// 初始化 peers 的 nextIndex 和 matchIndex
					rf.nextIndex = make([]int, len(rf.otherNodes))
					rf.matchIndex = make([]int, len(rf.otherNodes))
					// 为每个节点初始化 nextIndex 和 matchIndex，这里不考虑 Leader 重新选举的情况
					for i := range rf.otherNodes {
						rf.nextIndex[i] = 1
						rf.matchIndex[i] = 0
					}
					// 模拟客户端每 3s 发送一条 command
					go func() {
						i := 0
						for {
							i += 1
							// Leader 收到客户端的 command，封装成 LogEntry 并 append 到日志
							rf.log = append(rf.log, LogEntry{rf.currentTerm, i, fmt.Sprintf("user send: %d", i)})
							time.Sleep(3 * time.Second)
						}
					}()
				}
			case Leader:
				fmt.Println("This is Leader[%d]", rf.cid)
				// 定时心跳
				rf.broadcastHeartbeat()
				time.Sleep(100 * time.Microsecond)
			}
		}
	}()

}
