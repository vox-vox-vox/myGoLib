package geerpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"sync"
)

/*			   pending-map
requestA ---> |-------------|
requestB ---> | A,B,C,D,E   |
requestC ---> |             |--->A---> server
requestD ---> |             |
requestE ---> |_____________|

				pending-map
requestA <---  |-------------|
requestB block | A(removed)  |<---A<--- server
requestC block | B,C,D,E     |--->B---> server
requestD block |             |
requestE block |_____________|

*/

// Call 承载一次 RPC 调用所需要的信息
// call和body，header的数据结构关系
type Call struct {
	Seq           uint64 // 序号
	ServiceMethod string // 方法名称
	Error         error // 一次RPC调用产生的Error
	Args          interface{} // 方法参数
	Reply        interface{} //方法返回
	Done          chan *Call // todo:done-chan为什么要用 call作为element，用struct{}{}不行吗
}

// todo:为了支持异步调用，Call 结构体中添加了一个字段 Done，Done 的类型是 chan *Call，当调用结束时，会调用 call.done() 通知调用方。
func (call *Call) done() {
	call.Done <- call
}

type Client struct {
	cc       codec.Codec //cc 是消息的编解码器，和服务端类似，用来序列化将要发送出去的请求，以及反序列化接收到的响应。实质上相当于网络文件
	opt      *Option
	sending  sync.Mutex       // sending 是一个互斥锁，和服务端类似，为了保证请求的有序发送，即防止出现多个请求报文混淆。
	header   codec.Header     // header 是每个请求的消息头，header 只有在请求发送时才需要，而请求发送是互斥的，因此每个客户端只需要一个，声明在 Client 结构体中可以复用。
	mu       sync.Mutex       // protect following
	seq      uint64           // 用于给发送的请求编号，每个请求拥有唯一编号。
	pending  map[uint64]*Call // 存储未处理完的请求，键是编号，值是 Call 实例。
	closing  bool             // 用户主动关闭client
	shutdown bool             // server has told us to stop ，有错误发生
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing {
		return ErrShutdown
	}
	c.closing = true
	return c.cc.Close() // 关闭网络文件
}

func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.shutdown && !c.closing
}

// registerCall 将待发送请求call 放入pending中等待发送，并更新client.seq
func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}

// removeCall 根据seq，从pending中移除对应的call并返回，相当于消费
func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

// terminateCalls 服务端或客户端发生错误时调用，将shutdown置为true，并将错误信息通知所有pending状态的call
func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shutdown = true
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}


// receive
// call 不存在，可能是请求没有发送完整，或者因为其他原因被取消，但是服务端仍旧处理了。
// call 存在，但服务端处理出错，即 h.Error 不为空。
// call 存在，服务端处理正常，那么需要从 body 中读取 Reply 的值。
func (c *Client) receive() {
	var err error
	for err ==nil {
		var h codec.Header
		if err = c.cc.ReadHeader(&h);err!=nil{
			break
		}
		call := c.removeCall(h.Seq)
		switch{
		case call == nil:
			err = c.cc.ReadBody(nil)
		case h.Error!="":
			call.Error =fmt.Errorf(h.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			err = c.cc.ReadBody(call.Reply)
			if err!=nil{
				call.Error = errors.New("reading body"+ err.Error())
			}
			call.done()
		}
	}
	c.terminateCalls(err)
}

// send 发送一次RPC调用
func (c *Client) send(call *Call) {
	c.sending.Lock()
	defer c.sending.Unlock()

	// 记录一次调用
	seq,err:=c.registerCall(call)
	if err!=nil{
		call.Error = err
		call.done()
		return
	}

	// 给header赋值，由于sending是互斥的（加了锁）,可以不用新建新的header实例，每次只修改原header的值即可
	c.header.Seq=seq
	c.header.Error=""
	c.header.ServiceMethod=call.ServiceMethod

	if err := c.cc.Write(&c.header,call.Args);err != nil {
		// 如果写入失败了，移除这次调用
		call:=c.removeCall(seq)
		if call!=nil{
			call.Error=err
			call.done()
		}
		return
	}
}

/*
创建 Client 实例时，首先需要完成一开始的协议交换，即发送 Option 信息给服务端。协商好消息的编解码方式之后，再创建一个子协程调用 receive() 接收响应。
*/

func NewClient(conn net.Conn,opt *Option)(*Client,error){
	// 2. 发送option，向server完成协议交换
	f:=codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s",opt.CodecType)
		log.Println("rpc client: codec error:",err)
		return nil,err
	}

	if err:=json.NewEncoder(conn).Encode(opt);err!=nil{
		log.Println("rpc client :options error: ",err)
		_ =conn.Close()
		return nil,err
	}

	return newClientCodec(f(conn),opt),nil
}

func newClientCodec(cc codec.Codec,option *Option)*Client{
	client:=&Client{
		seq :1,
		cc:cc,
		opt:option,
		pending: make(map[uint64]*Call),
	}
	go client.receive()// client的receive是自动打开的，不暴露给用户
	return client
}

func parseOptions(opts ...*Option) (*Option, error) {
	// if opts is nil or pass nil as parameter
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

// Dial connects to an RPC server at the specified network address
func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	// close the connection if client is nil
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
}

// Go 非阻塞调用，返回一个Call，可以利用Call的Done信道判断方法是否调用完毕
/*
	call := client.Go( ... )
	# 新启动协程，异步等待
	go func(call *Call) {
		select {
			<-call.Done:
				# do something
			<-otherChan:
				# do something
		}
	}(call)

	otherFunc() # 不阻塞，继续执行其他函数。
*/
func (c *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	// 创建一次RPC调用
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	// 发送一次RPC调用
	c.send(call)
	return call
}

// Call 阻塞调用
func (c *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-c.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}