package geerpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yqchilde/gee-rpc/codec"
)

type Bar int

func (b Bar) Timeout(argv int, reply *int) error {
	time.Sleep(time.Second * 2)
	return nil
}

func startServer(addr chan string, port string, data interface{}) {
	_ = Register(data)
	l, _ := net.Listen("tcp", fmt.Sprintf(":%s", port))
	addr <- l.Addr().String()
	Accept(l)
}

func TestClient_dialTimeout(t *testing.T) {
	t.Parallel()
	l, _ := net.Listen("tcp", ":0")

	f := func(conn net.Conn, opt *Option) (client *Client, err error) {
		_ = conn.Close()
		time.Sleep(time.Second * 2)
		return nil, nil
	}
	t.Run("timeout", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", l.Addr().String(), &Option{ConnectTimeout: time.Second})
		assert.NotEqual(t, err != nil && strings.Contains(err.Error(), "connect timeout"), "expect a timeout error")
	})
	t.Run("0", func(t *testing.T) {
		_, err := dialTimeout(f, "tcp", l.Addr().String(), &Option{ConnectTimeout: 0})
		assert.Nil(t, err, "0 means no limit")
	})
}

func TestClient_Call(t *testing.T) {
	t.Parallel()
	addrCh := make(chan string)

	go startServer(addrCh, "8080", new(Bar))
	addr := <-addrCh
	time.Sleep(time.Second)
	t.Run("client timeout", func(t *testing.T) {
		client, _ := Dial("tcp", addr)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer func() { cancel() }()
		var reply int
		assert.NotEqual(t, client.IsAvailable() == true, "client is closed")
		err := client.Call(ctx, "Bar.Timeout", 1, &reply)
		client.Close()
		assert.NotEqual(t, err != nil && strings.Contains(err.Error(), ctx.Err().Error()), "expect a timeout error")

	})
	t.Run("server handle timeout", func(t *testing.T) {
		client, _ := Dial("tcp", addr, &Option{
			MagicNumber:   MagicNumber,
			CodecType:     codec.GobType,
			HandleTimeout: time.Second,
		})
		var reply int
		err := client.Call(context.Background(), "Bar.Timeout", 1, &reply)
		client.Close()
		assert.NotEqual(t, err != nil && strings.Contains(err.Error(), "handle timeout"), "expect a timeout error")
	})
}

func TestXDial(t *testing.T) {
	t.Parallel()
	ch := make(chan struct{})

	t.Run("XDial with http protocol", func(t *testing.T) {
		addr := ":9999"
		go func() {
			l, err := net.Listen("tcp", addr)
			if err != nil {
				t.Fatal("failed to listen tcp")
			}
			ch <- struct{}{}
			HandleHTTP()
			_ = http.Serve(l, nil)
		}()
		<-ch
		_, err := XDial("http@" + addr)
		assert.Nil(t, err, "failed to connect http")
	})
	t.Run("XDial with tcp protocol", func(t *testing.T) {
		addr := ":9998"
		go func() {
			l, err := net.Listen("tcp", addr)
			if err != nil {
				t.Fatal("failed to listen tcp")
			}
			ch <- struct{}{}
			Accept(l)
		}()
		<-ch
		_, err := XDial("tcp@" + addr)
		assert.Nil(t, err, "failed to connect tcp")
	})
	t.Run("XDial parts more than 2", func(t *testing.T) {
		_, err := XDial("tcp@127.0.0.1@9999")
		assert.NotNil(t, err, "failed to connect tcp")
	})
}
