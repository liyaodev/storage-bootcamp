FROM liyaodev/base-cpu-u18-py3.8:v1.0.0
LABEL maintainer=liyaodev

# Install Go 1.18.1 to /usr/local/go
ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV PATH $GOPATH/bin:$GOROOT/bin:$PATH
RUN curl -s https://dl.google.com/go/go1.18.1.linux-amd64.tar.gz | tar -v -C /usr/local -xz && \
    mkdir -p "$GOPATH/src" "$GOPATH/bin"

# Config Go Env
RUN go install -v github.com/ramya-rao-a/go-outline@latest && \
    go install -v golang.org/x/tools/gopls@latest && \
    go install -v github.com/uudashr/gopkgs/v2/cmd/gopkgs@latest && \
    go install -v github.com/go-delve/delve/cmd/dlv@latest && \
    go install -v honnef.co/go/tools/cmd/staticcheck@latest && \
    go clean --modcache

RUN go env -w GO111MODULE=on && \
    go env -w GOPROXY=https://goproxy.cn,direct

RUN echo 'root:root' | chpasswd

# 构建Tini的多服务容器
RUN wget -O /tini https://github.com/krallin/tini/releases/download/v0.19.0/tini && \
    chmod +x /tini
ENTRYPOINT ["/tini", "--"]

WORKDIR /www/server

CMD ["tail", "-f", "/dev/null"]
EXPOSE 8888
