package geerpc

import (
	"encoding/json"
	"errors"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

/*
geerpc 需要协商的唯一内容是消息的编码方式，放入Options中承载。Options固定用Json编码，解码出Options后，得到CodeType，再利用CodeType去解码后面的Header和Body

| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|

一次连接中，Option 固定在报文的最开始，Header 和 Body 可以有多个
| Option | Header1 | Body1 | Header2 | Body2 | ...

Header
*/

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

//Server is a RPC server
type Server struct{
	serviceMap sync.Map
}

// NewServer returns a new Server
func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept() // 可以接受多个tcp连接
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go s.ServeConn(conn)// 开启一个goroutine 处理一次client的链接
	}
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

// ServeConn 解析connection连接
// 1. 反序列化得到Option实例 2. 检查MagicNumber 和 CodeType 3.根据CodeType得到编码器
func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	//todo: _是什么意思
	defer func() { _ = conn.Close() }()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc Server:options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server:invalid magicNumber: %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server:code type %s", opt.CodecType)
		return
	}
	s.serveCodec(f(conn))

}

var invalidRequest = struct{}{}

func (s *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex) // 回复请求的报文必须是逐个发送的，这里用锁保证
	wg := new(sync.WaitGroup) // 一个waitGroup对象可以等待一组协程结束

	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break // client 调用 close 时，发送FIN报文，server读到FIN报文表现在应用端为error=EOF，此时request=0，跳出循环，说明client已经关闭。
			}
			req.header.Error = err.Error()
			s.sendResponse(cc,req.header,invalidRequest,sending)
		}
		wg.Add(1)
		go s.handleRequest(cc,req,sending,wg)
	}
	wg.Wait()
	_ = cc.Close()
}

// request
/*
	Header:
	service:
	method:
	argv,replyv
*/
type request struct {
	header       *codec.Header // header of request
	svc          *service
	mtype        *methodType
	argv, replyv reflect.Value // argv and replyv of request
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

// readRequest 解析得到request
func (s *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{header: h}
	// methodType

	req.svc, req.mtype, err = s.findService(h.ServiceMethod)
	if err != nil {
		return nil, err
	}
	// 新建argv，replyv实例
	req.argv=req.mtype.newArgv()
	req.replyv=req.mtype.newReplyv()

	// make sure that argvi is a pointer, ReadBody need a pointer as parameter
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	// 传入指针的interface
	if err = cc.ReadBody(argvi);err!=nil{
		log.Println("rpc server: read argv err:",err)
	}

	return req,nil
}

func(s *Server) handleRequest(cc codec.Codec,req *request,sending *sync.Mutex,wg *sync.WaitGroup){
	// TODO, should call registered rpc methods to get the right replyv
	// day 1, just print argv and send a hello message
	defer wg.Done()
	err := req.svc.call(req.mtype, req.argv, req.replyv)
	if err != nil {
		req.header.Error = err.Error()
		s.sendResponse(cc, req.header, invalidRequest, sending)
		return
	}
	s.sendResponse(cc, req.header, req.replyv.Interface(), sending)
}

func(s *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{},sending *sync.Mutex){
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h,body); err!=nil{
		log.Println("rpc server :write response error:",err)
	}
}

// ===================注册service====================

// Register publishes in the server the set of methods of the
func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
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




















