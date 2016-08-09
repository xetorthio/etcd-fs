package main

import (
	"bytes"
	"log"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"

	etcd "github.com/coreos/etcd/client"
)

type etcdFile struct {
	kv   etcd.KeysAPI
	path string
}

func NewEtcdFile(kv etcd.KeysAPI, path string) nodefs.File {
	file := new(etcdFile)
	file.kv = kv
	file.path = path
	return file
}

func (f *etcdFile) SetInode(*nodefs.Inode) {
}
func (f *etcdFile) InnerFile() nodefs.File {
	return nil
}

func (f *etcdFile) String() string {
	return "etcdFile"
}

func (f *etcdFile) Read(buf []byte, off int64) (fuse.ReadResult, fuse.Status) {
	res, err := f.kv.Get(
		ctx(),
		f.path,
		&etcd.GetOptions{},
	)

	if err != nil {
		log.Println("Error:", err)
		return nil, fuse.EIO
	}

	end := int(off) + int(len(buf))
	if end > len(res.Node.Value) {
		end = len(res.Node.Value)
	}

	data := []byte(res.Node.Value)
	return fuse.ReadResultData(data[off:end]), fuse.OK
}

func (f *etcdFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	res, err := f.kv.Get(
		ctx(),
		f.path,
		&etcd.GetOptions{},
	)

	if err != nil {
		log.Println("Error:", err)
		return 0, fuse.EIO
	}

	originalValue := []byte(res.Node.Value)
	leftChunk := originalValue[:off]
	end := int(off) + int(len(data))

	var rightChunk []byte
	if end > len(res.Node.Value) {
		rightChunk = []byte{}
	} else {
		rightChunk = originalValue[int(off)+int(len(data)):]
	}

	newValue := bytes.NewBuffer(leftChunk)
	newValue.Grow(len(data) + len(rightChunk))
	newValue.Write(data)
	newValue.Write(rightChunk)

	_, err = f.kv.Set(
		ctx(),
		f.path,
		newValue.String(),
		&etcd.SetOptions{
			PrevExist: etcd.PrevExist,
		},
	)

	if err != nil {
		log.Println("Error:", err)
		return 0, fuse.EIO
	}

	return uint32(len(data)), fuse.OK
}

func (f *etcdFile) Flush() fuse.Status {
	return fuse.OK
}

func (f *etcdFile) Release() {
}

func (f *etcdFile) GetAttr(out *fuse.Attr) fuse.Status {
	res, err := f.kv.Get(
		ctx(),
		f.path,
		&etcd.GetOptions{},
	)

	if err != nil {
		log.Println("Error:", err)
		return fuse.EIO
	}

	out.Mode = fuse.S_IFREG | 0666
	out.Size = uint64(len(res.Node.Value))
	return fuse.OK
}

func (f *etcdFile) Fsync(flags int) (code fuse.Status) {
	return fuse.OK
}

func (f *etcdFile) Utimens(atime *time.Time, mtime *time.Time) fuse.Status {
	return fuse.ENOSYS
}

func (f *etcdFile) Truncate(size uint64) fuse.Status {
	res, err := f.kv.Get(
		ctx(),
		f.path,
		&etcd.GetOptions{},
	)

	if err != nil {
		log.Println("Error:", err)
		return fuse.EIO
	}

	if uint64(len(res.Node.Value)) < size {
		log.Printf("Error: Truncating file with size %v to %v\n", len(res.Node.Value), size)
		return fuse.EIO
	}

	_, err = f.kv.Set(
		ctx(),
		f.path,
		res.Node.Value[:size],
		&etcd.SetOptions{
			PrevExist: etcd.PrevExist,
		},
	)

	if err != nil {
		log.Println("Error:", err)
		return fuse.EIO
	}

	return fuse.OK
}

func (f *etcdFile) Chown(uid uint32, gid uint32) fuse.Status {
	return fuse.ENOSYS
}

func (f *etcdFile) Chmod(perms uint32) fuse.Status {
	return fuse.ENOSYS
}

func (f *etcdFile) Allocate(off uint64, size uint64, mode uint32) (code fuse.Status) {
	return fuse.OK
}
