package main

import (
	etcd "github.com/coreos/etcd/client"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flag.Parse()
	if len(flag.Args()) < 2 {
		log.Fatal("Usage:\n  etcd-fs MOUNTPOINT ETCDENDPOINT[s]")
	}

	etcdClient, err := etcd.New(etcd.Config{
		Endpoints: flag.Args()[1:],
		Transport: etcd.DefaultTransport,
	})

	if err != nil {
		log.Fatal(err)
	}

	etcdFs := EtcdFs{
		FileSystem: pathfs.NewDefaultFileSystem(),
		Client:     etcdClient,
	}

	nfs := pathfs.NewPathNodeFs(&etcdFs, nil)

	server, _, err := nodefs.MountRoot(flag.Arg(0), nfs.Root(), nil)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer func() { //catch any panic
			if r := recover(); r != nil {
				log.Println(r)
				sigs <- syscall.SIGINT
			}
		}()

		server.Serve()
	}()

	<-sigs
	server.Unmount()
}
