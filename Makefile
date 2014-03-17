export GOPATH=$(shell pwd)

test:
	-docker -H "localhost:4243" run -d -p 8001:8001 -p 4001:4001 --name etcd-node1 coreos/etcd -name etcd-node1
	go get github.com/franela/goblin
	go get github.com/coreos/go-etcd/etcd
	go get github.com/hanwen/go-fuse/fuse
	go test -v etcdfs

build:
	go build etcdfs.go
