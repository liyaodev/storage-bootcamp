
## 服务发现实例

### 环境构建

```shell
go mod init

go get go.etcd.io/etcd/clientv3
go get github.com/coreos/etcd/mvcc/mvccpb
go get google.golang.org/grpc@v1.26.0
```

### 运行

```shell
# 启动客户端，监听和维护服务列表
go run client.go

# 启动服务端，注册和保持服务链接
go run server.go  
```