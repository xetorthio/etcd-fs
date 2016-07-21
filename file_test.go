package main

import (
	etcdm "github.com/coreos/go-etcd/etcd"
	. "github.com/franela/goblin"
	"testing"
	//  "fmt"
	"io/ioutil"
	"os"
)

func TestNodeFs(t *testing.T) {
	g := Goblin(t)

	g.Describe("File", func() {
		var etcd *etcdm.Client
		var fs testEtcdFsMount

		g.Before(func() {
			etcd = etcdm.NewClient([]string{testEtcdEndpoint})
		})

		g.BeforeEach(func() {
			etcd.RawDelete("/test", true, true)
			etcd.SetDir("/test", 0)
			fs = NewTestEtcdFsMount()
		})

		g.AfterEach(func() {
			fs.Unmount()
		})

		g.Describe("Open", func() {
			g.It("Should be supported", func() {
				if _, e := etcd.Set("/test/foo", "bar", 0); e != nil {
					g.Fail(e)
				}

				file, err := os.Open(fs.Path() + "/test/foo")

				if err != nil {
					g.Fail(err)
				}

				file.Close()
			})
		})
		g.Describe("Create", func() {
			g.It("Should be supported", func() {
				file, err := os.Create(fs.Path() + "/test/bar")

				if err != nil {
					g.Fail(err)
				}
				file.Close()

				if _, er := etcd.Get("/test/bar", false, false); er != nil {
					g.Fail(er)
				}
			})
		})
		g.Describe("Delete", func() {
			g.It("Should be supported", func() {
				etcd.Set("/test/barfoo", "lala", 0)

				err := os.Remove(fs.Path() + "/test/barfoo")

				if err != nil {
					g.Fail(err)
				}

				if _, er := etcd.Get("/test/barfoo", false, false); er == nil {
					g.Fail("The key [/test/barfoo] should not exist")
				}
			})
		})
		g.Describe("Read", func() {
			g.It("Should be supported", func() {
				etcd.Set("/test/bar", "foo", 0)

				data, err := ioutil.ReadFile(fs.Path() + "/test/bar")

				if err != nil {
					g.Fail(err)
				}

				g.Assert(string(data)).Equal("foo")
			})
		})
		g.Describe("Write", func() {
			g.It("Should be supported", func() {
				if err := ioutil.WriteFile(fs.Path()+"/test/foobar", []byte("hello world"), 0666); err != nil {
					g.Fail(err)
				}

				res, err := etcd.Get("/test/foobar", false, false)

				if err != nil {
					g.Fail(err)
				}

				g.Assert(res.Node.Value).Equal("hello world")
			})
		})
	})
}
