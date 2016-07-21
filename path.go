package main

import (
	"log"
	"strings"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/client"
)

// EtcdFs is filesystem state
type EtcdFs struct {
	kvapi etcd.KeysAPI

	pathfs.FileSystem
	etcd.Client
}

func (fs *EtcdFs) kv() etcd.KeysAPI {
	if fs.kvapi == nil {
		fs.kvapi = etcd.NewKeysAPI(fs)
	}

	return fs.kvapi
}

func ctx() context.Context {

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	return ctx
}

func (me *EtcdFs) Unlink(name string, context *fuse.Context) (code fuse.Status) {
	if name == "" {
		return fuse.OK
	}

	_, err := me.kv().Delete(
		ctx(),
		name,
		&etcd.DeleteOptions{Recursive: false},
	)

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

	_, err := me.kv().Delete(
		ctx(),
		name,
		&etcd.DeleteOptions{
			Recursive: true,
			Dir:       true,
		},
	)

	if err != nil {
		log.Println(err)
		return fuse.ENOENT
	}

	return fuse.OK
}

func (me *EtcdFs) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	_, err := me.kv().Set(
		ctx(),
		name,
		"",
		&etcd.SetOptions{
			PrevExist: etcd.PrevNoExist,
		},
	)

	if err != nil {
		log.Println("Create Error:", err)
		return nil, fuse.ENOENT
	}

	return NewEtcdFile(me.kv(), name), fuse.OK
}

func (me *EtcdFs) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	if name == "" {
		return fuse.OK
	}

	_, err := me.kv().Set(
		ctx(),
		name,
		"",
		&etcd.SetOptions{
			Dir: true,
		},
	)

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

	res, err := me.kv().Get(
		ctx(),
		name,
		&etcd.GetOptions{},
	)

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
	res, err := me.kv().Get(
		ctx(),
		name,
		&etcd.GetOptions{},
	)

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
	_, err := me.kv().Get(
		ctx(),
		name,
		&etcd.GetOptions{},
	)

	if err != nil {
		log.Println("Open Error:", err)
		return nil, fuse.ENOENT
	}

	return NewEtcdFile(me.kv(), name), fuse.OK
}

func (me *EtcdFs) Rename(oldName string, newName string, context *fuse.Context) (code fuse.Status) {

	getRes, err := me.kv().Get(
		ctx(),
		oldName,
		&etcd.GetOptions{},
	)

	if err != nil {
		log.Println("Open Error:", err)
		return fuse.ENOENT
	}

	_, err = me.kv().Set(
		ctx(),
		newName,
		getRes.Node.Value,
		&etcd.SetOptions{
			Dir: false,
		},
	)

	if err != nil {
		log.Println("Open Error:", err)
		return fuse.ENOENT
	}

	_, err = me.kv().Delete(
		ctx(),
		oldName,
		&etcd.DeleteOptions{
			Dir: false,
		},
	)

	if err != nil {
		log.Println("Open Error:", err)
		return fuse.ENOENT
	}

	return fuse.OK

}
