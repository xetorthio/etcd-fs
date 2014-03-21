package etcdfs

import (
	"log"
	"strings"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"

	"github.com/coreos/go-etcd/etcd"
)

type EtcdFs struct {
	pathfs.FileSystem
	EtcdEndpoint string
}

func (me *EtcdFs) NewEtcdClient() *etcd.Client {
	return etcd.NewClient([]string{me.EtcdEndpoint})
}

func (me *EtcdFs) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	if name == "" {
		return fuse.OK
	}

	_, err := me.NewEtcdClient().Delete(name, false)

	if err != nil {
		log.Println(err)
		return fuse.ENOENT
	}

	return fuse.OK
}

func (me *EtcdFs) Rmdir(name string, context *fuse.Context) (code fuse.Status) {
	if name == "" {
		return fuse.OK
	}

	_, err := me.NewEtcdClient().RawDelete(name, true, true)

	if err != nil {
		log.Println(err)
		return fuse.ENOENT
	}

	return fuse.OK
}

func (me *EtcdFs) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	_, err := me.NewEtcdClient().Set(name, "", 0)

	if err != nil {
		log.Println("Create Error:", err)
		return nil, fuse.ENOENT
	}

	return NewEtcdFile(me.NewEtcdClient(), name), fuse.OK
}

func (me *EtcdFs) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	if name == "" {
		return fuse.OK
	}

	_, err := me.NewEtcdClient().CreateDir(name, 0)

	if err != nil {
		log.Println(err)
		return fuse.ENOENT
	}

	return fuse.OK
}

func (me *EtcdFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	if name == "" {
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0666,
		}, fuse.OK
	}

	res, err := me.NewEtcdClient().Get(name, false, false)

	if err != nil {
		return nil, fuse.ENOENT
	}

	var attr fuse.Attr

	if res.Node.Dir {
		attr = fuse.Attr{
			Mode: fuse.S_IFDIR | 0666,
		}
	} else {
		attr = fuse.Attr{
			Mode: fuse.S_IFREG | 0666, Size: uint64(len(res.Node.Value)),
		}
	}

	return &attr, fuse.OK
}

func (me *EtcdFs) OpenDir(name string, context *fuse.Context) (c []fuse.DirEntry, code fuse.Status) {
	res, err := me.NewEtcdClient().Get(name, false, false)

	if err != nil {
		log.Println("OpenDir Error:", err)
		return nil, fuse.ENOENT
	}

	entries := []fuse.DirEntry{}

	for _, e := range res.Node.Nodes {
		chunks := strings.Split(e.Key, "/")
		file := chunks[len(chunks)-1]
		if e.Dir {
			entries = append(entries, fuse.DirEntry{Name: file, Mode: fuse.S_IFDIR})
		} else {
			entries = append(entries, fuse.DirEntry{Name: file, Mode: fuse.S_IFREG})
		}
	}

	return entries, fuse.OK
}

func (me *EtcdFs) Open(name string, flags uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	_, err := me.NewEtcdClient().Get(name, false, false)

	if err != nil {
		log.Println("Open Error:", err)
		return nil, fuse.ENOENT
	}

	return NewEtcdFile(me.NewEtcdClient(), name), fuse.OK
}
