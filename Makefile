export GOPATH=$(shell pwd)

test:
	-docker run -d -p 8001:8001 -p 4001:4001 coreos/etcd -name etcd-node1
	go get github.com/franela/goblin
	go get github.com/coreos/go-etcd/etcd
	go get github.com/hanwen/go-fuse/fuse
	go test -v etcdfs

install:
	sudo apt-get install -qq fuse
	sudo modprobe fuse

build:
	go build etcdfs.go
