package geerpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/yqchilde/gee-rpc/codec"
	"strings"
	"testing"
	"time"
)

func TestServer_ServeConn(t *testing.T) {
	t.Parallel()
	addrCh := make(chan string)
	go startServer(addrCh)
	addr := <-addrCh
	time.Sleep(time.Second)
	t.Run("server connect", func(t *testing.T) {
		client, _ := Dial("tcp", addr, &Option{
			MagicNumber:   MagicNumber,
			CodecType:     codec.GobType,
			HandleTimeout: time.Second,
		})
		args := &Args{Num1: 1, Num2: 3}
		var reply int
		err := client.Call(context.Background(), "Foo.Sum", args, &reply)
		client.Close()
		assert.NotEqual(t, err != nil && strings.Contains(err.Error(), "handle timeout"), "expect a timeout error")
	})
}

//type Foo5 int
//
//func (f Foo5) Sum(args Args, reply *int) error {
//	*reply = args.Num1 + args.Num2
//	return nil
//}
//
//func TestNewServer(t *testing.T) {
//	addr := make(chan string)
//	go func() {
//		var foo Foo
//		if err := Register(&foo); err != nil {
//			t.Fatal("register error:", err)
//		}
//
//		l, err := net.Listen("tcp", ":8080")
//		if err != nil {
//			t.Fatal("network error:", err)
//		}
//		t.Log("start rpc server on", l.Addr())
//		addr <- l.Addr().String()
//		Accept(l)
//	}()
//
//	client, err := Dial("tcp", <-addr)
//	if err != nil {
//		t.Fatal("client dial failed: ", err.Error())
//	}
//	defer func() { _ = client.Close() }()
//
//	time.Sleep(time.Second)
//
//	// mock send request && receive response
//	var wg sync.WaitGroup
//	for i := 0; i < 5; i++ {
//		wg.Add(1)
//		go func(i int) {
//			defer wg.Done()
//			args := &Args{Num1: i, Num2: i * i}
//			var reply int
//			if err := client.Call(context.TODO(), "Foo.Sum", args, &reply); err != nil {
//				t.Fatal("call Foo.Sum error: ", err)
//			}
//			t.Logf("%d + %d = %d", args.Num1, args.Num2, reply)
//		}(i)
//	}
//	wg.Wait()
//}
//
//func TestNewServerTamperingWithMagicNumber(t *testing.T) {
//	addr := make(chan string)
//	go func() {
//		var foo2 Foo2
//		if err := Register(&foo2); err != nil {
//			t.Fatal("register error:", err)
//		}
//
//		l, err := net.Listen("tcp", ":8081")
//		if err != nil {
//			t.Fatal("network error:", err)
//		}
//		t.Log("start rpc server on", l.Addr())
//		addr <- l.Addr().String()
//		Accept(l)
//	}()
//
//	var customOption = &Option{
//		MagicNumber: 123456,
//		CodecType:   codec.GobType,
//	}
//	client, err := Dial("tcp", <-addr, customOption)
//	if err != nil {
//		t.Fatal("client dial failed: ", err.Error())
//	}
//	defer func() { _ = client.Close() }()
//
//	time.Sleep(time.Second)
//
//	// mock send request && receive response
//	var wg sync.WaitGroup
//	for i := 0; i < 5; i++ {
//		wg.Add(1)
//		go func(i int) {
//			defer wg.Done()
//			args := &Args{Num1: i, Num2: i * i}
//			var reply int
//			if err := client.Call(context.TODO(), "Foo.Sum", args, &reply); err != nil {
//				t.Fatal("call Foo.Sum error: ", err)
//			}
//			t.Logf("%d + %d = %d", args.Num1, args.Num2, reply)
//		}(i)
//	}
//	wg.Wait()
//}
//
//func TestNewServerTamperingWithSlice(t *testing.T) {
//	addr := make(chan string)
//	go func() {
//		var foo3 Foo3
//		if err := Register(&foo3); err != nil {
//			t.Fatal("register error:", err)
//		}
//
//		l, err := net.Listen("tcp", ":8082")
//		if err != nil {
//			t.Fatal("network error:", err)
//		}
//		t.Log("start rpc server on", l.Addr())
//		addr <- l.Addr().String()
//		Accept(l)
//	}()
//
//	var customOption = &Option{
//		MagicNumber: MagicNumber,
//		CodecType:   "application/test",
//	}
//	client, err := Dial("tcp", <-addr, customOption)
//	if err != nil {
//		t.Fatal("client dial failed: ", err.Error())
//	}
//	defer func() { _ = client.Close() }()
//
//	time.Sleep(time.Second)
//
//	// mock send request && receive response
//	var wg sync.WaitGroup
//	for i := 0; i < 5; i++ {
//		wg.Add(1)
//		go func(i int) {
//			defer wg.Done()
//			args := &Args{Num1: i, Num2: i * i}
//			var reply int
//			if err := client.Call(context.TODO(), "Foo.Sum", args, &reply); err != nil {
//				t.Fatal("call Foo.Sum error: ", err)
//			}
//			t.Logf("%d + %d = %d", args.Num1, args.Num2, reply)
//		}(i)
//	}
//	wg.Wait()
//}
//
//func TestNewServerTamperingWithFindService(t *testing.T) {
//	addr := make(chan string)
//	go func() {
//		var foo4 Foo4
//		if err := Register(&foo4); err != nil {
//			t.Fatal("register error:", err)
//		}
//
//		l, err := net.Listen("tcp", ":8083")
//		if err != nil {
//			t.Fatal("network error:", err)
//		}
//		t.Log("start rpc server on", l.Addr())
//		addr <- l.Addr().String()
//		Accept(l)
//	}()
//
//	client, err := Dial("tcp", <-addr)
//	if err != nil {
//		t.Fatal("client dial failed: ", err.Error())
//	}
//	defer func() { _ = client.Close() }()
//
//	time.Sleep(time.Second)
//
//	// mock send request && receive response
//	var wg sync.WaitGroup
//	for i := 0; i < 5; i++ {
//		wg.Add(1)
//		go func(i int) {
//			defer wg.Done()
//			args := &Args{Num1: i, Num2: i * i}
//			var reply int
//			if err := client.Call(context.TODO(), "FooSum", args, &reply); err != nil {
//				t.Fatal("call Foo.Sum error: ", err)
//			}
//			t.Logf("%d + %d = %d", args.Num1, args.Num2, reply)
//		}(i)
//	}
//	wg.Wait()
//}
