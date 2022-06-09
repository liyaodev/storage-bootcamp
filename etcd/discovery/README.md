
## 服务发现实例

### 依赖安装

```shell
go get go.etcd.io/etcd/clientv3
go get github.com/coreos/etcd/mvcc/mvccpb
go get google.golang.org/grpc@v1.26.0
```

### 运行

```shell
# 启动客户端，监听和维护服务列表
go run client/main.go

# 启动服务端，注册和保持服务链接
go run server/main.go  
```
