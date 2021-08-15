package geerpc

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"

	"github.com/yqchilde/gee-rpc/codec"
)

const MagicNumber = 0x5add9a7

type Option struct {
	MagicNumber int        // 标记这是一个rpc请求
	CodecType   codec.Type // 客户端可以选择不同的编解码器来编码正文
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

// Accept 接受侦听器上的每个接入连接，并且发送连接请求
func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go s.ServeConn(conn)
	}
}

// ServeConn 在单个连接上运行服务器
// 程序阻塞，服务连接直到客户端断开
func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	s.ServeCodec(f(conn))
}

var invalidRequest = struct{}{}

// ServeCodec 服务端编解码并执行请求返回响应
func (s *Server) ServeCodec(c codec.Codec) {
	var sending = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	for {
		req, err := s.readRequest(c)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			s.sendResponse(c, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(c, req, sending, wg)
	}
	wg.Wait()
	_ = c.Close()
}

type request struct {
	h            *codec.Header // 请求头
	argv, replyv reflect.Value // 请求参数和请求应答参数
	mtype        *methodType   // 请求方法
	svc          *service      // 请求服务
}

func (s *Server) readRequest(c codec.Codec) (*request, error) {
	h, err := s.readRequestHeader(c)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	req.svc, req.mtype, err = s.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()

	argvi := req.argv.Interface()
	if req.argv.Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	if err = c.ReadBody(argvi); err != nil {
		log.Println("rpc server: read argv err:", err)
		return req, err
	}
	return req, nil
}

func (s *Server) readRequestHeader(c codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := c.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) sendResponse(c codec.Codec, header *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := c.Write(header, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (s *Server) handleRequest(c codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	err := req.svc.call(req.mtype, req.argv, req.replyv)
	if err != nil {
		req.h.Error = err.Error()
		s.sendResponse(c, req.h, invalidRequest, sending)
		return
	}
	s.sendResponse(c, req.h, req.replyv.Interface(), sending)
}

// Accept 服务端开始接受请求
func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func (s *Server) Register(rcvr interface{}) error {
	svc := newService(rcvr)
	if _, dup := s.serviceMap.LoadOrStore(svc.name, svc); dup {
		return errors.New("rpc: service already defined: " + svc.name)
	}
	return nil
}

func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}

func (s *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := s.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}
