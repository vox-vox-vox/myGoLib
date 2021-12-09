package server

import (
	"log"
	"net"
	"net/rpc"
)

type HelloService struct{}

func (helloservice *HelloService)Hello(request string,reply *string) error {
	*reply="hello"+request
	return nil
}

func StartServerAndRegister(){
	err := rpc.RegisterName("HelloService", new(HelloService)) // 注册service
	if err != nil {
		return
	}
	listener,err:=net.Listen("tcp","localhost:8011")
	if err!=nil{
		log.Fatal("ListenTCP error",err)
	}
	conn, err :=listener.Accept()
	if err != nil {
		log.Fatal("Accept error:", err)
	}
	rpc.ServeConn(conn)
}
