
## 负载均衡实例

### protobuf

```shell
apt-get update && apt-get install protobuf-compiler -y
go install -v github.com/golang/protobuf/proto@latest
go install -v google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

protoc --go_out=. --go-grpc_out=. proto/hello.proto
```

### 运行

```shell
# 启动客户端，监听和维护服务列表
go run client.go

# 启动服务端，注册和保持服务链接
go run server.go  
```