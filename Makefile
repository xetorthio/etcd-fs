export GOPATH=$(shell pwd)

test:
	go get github.com/franela/goblin
	go get github.com/coreos/go-etcd/etcd
	go get github.com/hanwen/go-fuse/fuse
	go test -v etcdfs

build:
	go build etcdfs.go
