export GOPATH=$(shell pwd)

test:
	go get github.com/franela/goblin
	go get github.com/coreos/go-etcd/etcd
	go get github.com/hanwen/go-fuse/fuse
	go test -v etcdfs

install:
	sudo apt-get install -qq fuse
	sudo modprobe fuse
	git clone https://github.com/coreos/etcd
	cd etcd ; ./build ; ./bin/etcd &

build:
	go build etcdfs.go
