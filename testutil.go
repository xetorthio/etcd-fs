package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	etcd "github.com/coreos/etcd/client"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

const (
	testTtl          = 100 * time.Millisecond
	testVerbose      = false
	testEtcdEndpoint = "http://localhost:4001"
)

type testEtcdFsMount struct {
	path  string
	state *fuse.Server
}

func (me testEtcdFsMount) Path() string {
	return me.path
}

func (me testEtcdFsMount) Unmount() {
	err := me.state.Unmount()

	if err != nil {
		fmt.Printf("Unmount failed: %v\n", err)
	}

	os.RemoveAll(me.path)
}

func NewTestEtcdFsMount() testEtcdFsMount {
	t := testEtcdFsMount{}

	var err error
	t.path, err = ioutil.TempDir("", "etcd-fs")

	if err != nil {
		fmt.Printf("Temdir fail: %v\n", err)
	}

	etcdClient, err := etcd.New(etcd.Config{
		Endpoints: []string{testEtcdEndpoint},
		Transport: etcd.DefaultTransport,
	})

	if err != nil {
		fmt.Printf("Connect to etcd fail: %v\n", err)
	}

	etcdFs := EtcdFs{FileSystem: pathfs.NewDefaultFileSystem(), Client: etcdClient}

	nfs := pathfs.NewPathNodeFs(&etcdFs, nil)

	connector := nodefs.NewFileSystemConnector(nfs.Root(), &nodefs.Options{EntryTimeout: testTtl, AttrTimeout: testTtl, NegativeTimeout: 0.0})
	connector.SetDebug(testVerbose)

	t.state, err = fuse.NewServer(fuse.NewRawFileSystem(connector.RawFS()), t.path, &fuse.MountOptions{SingleThreaded: true})
	if err != nil {
		fmt.Println("NewServer:", err)
	}

	t.state.SetDebug(testVerbose)

	// Unthreaded, but in background.
	go t.state.Serve()

	t.state.WaitMount()

	return t
}
