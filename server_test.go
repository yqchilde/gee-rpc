package geerpc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yqchilde/gee-rpc/codec"
)

func TestServer_ServeConn(t *testing.T) {
	t.Parallel()
	addrCh := make(chan string)
	go startServer(addrCh, "8081", new(Foo))
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
