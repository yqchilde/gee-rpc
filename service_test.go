package geerpc

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Foo int
type Foo2 int
type Foo3 int
type Foo4 int

type Args struct {
	Num1, Num2 int
}

type args struct {
	Num1, Num2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo2) Sum(args Args, reply *int, tmp int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo2) SumArgPointer(args *Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo2) SumRetMap(args *Args, reply map[string]map[string]int) error {
	d := make(map[string]int, 1)
	d["result"] = args.Num1 + args.Num2
	reply["result"] = d
	return nil
}

func (f Foo2) SumRetSlice(args *Args, reply [][]int) error {
	reply[0][0] = args.Num1 + args.Num2
	return nil
}

func (f Foo3) Sum(args Args, reply *int) *int {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo4) Sums(args args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	assert.Equal(t, len(s.method), 1, "wrong service Method, expect 1, bug got %d", len(s.method))
	mType := s.method["Sum"]
	assert.NotNil(t, mType, "wrong Method, Sum shouldn't nil")
	mType2 := s.method["sum"]
	assert.Nil(t, mType2, "wrong Method, sum should nil")

	var foo2 Foo2
	_ = newService(&foo2)

	var foo3 Foo3
	_ = newService(&foo3)

	var foo4 Foo4
	_ = newService(&foo4)
}

func TestMethodCall(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["Sum"]

	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{Num1: 1, Num2: 3}))
	err := s.call(mType, argv, replyv)
	assert.NotEqual(t, err == nil && *replyv.Interface().(*int) == 4 && mType.numCalls == 1, "failed to call Foo.Sum")

	var foo2 Foo2
	s2 := newService(&foo2)
	s2.method["SumArgPointer"].newArgv()
	s2.method["SumRetMap"].newReplyv()
	s2.method["SumRetSlice"].newReplyv()
}
