package common

import "time"

var (
	dialTimeout  = 5 * time.Second
	leaseTimeout = int64(3)
	etcdPrefix   = "/etcd/discovery"
)
