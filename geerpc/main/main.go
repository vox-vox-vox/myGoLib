package main

import (
	"fmt"
	"geerpc"
	"log"
	"net"
	"os"
	"time"
)

//func startServer(addr chan string){
//	l,err := net.Listen("tcp","localhost:8002")
//	if err!=nil{
//		log.Fatal("network error:",err)
//	}
//	log.Println("start rpc server on",l.Addr())
//	addr <- l.Addr().String()
//	geerpc.Accept(l)
//}

//func main() {
//	addr := make(chan string)
//	go startServer(addr)
//
//	// in fact, following code is like a simple geerpc client
//	conn, _ := net.Dial("tcp", <-addr)
//	defer func() { _ = conn.Close() }()
//
//	time.Sleep(time.Second)
//	// send options
//	_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)
//	cc := codec.NewGobCodec(conn)
//	// send request & receive response
//	for i := 0; i < 5; i++ {
//		h := &codec.Header{
//			ServiceMethod: "Foo.Sum",
//			Seq:           uint64(i),
//		}
//		_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
//		_ = cc.ReadHeader(h)
//		var reply string
//		_ = cc.ReadBody(&reply)
//		log.Println("reply:", reply)
//	}
//}

// Foo 是一个test实例，用于服务注册
type Foo int
type Input struct {
	Input1 int
	Input2 int
}
// Sum 形式一定要遵循 func(arg,&reply) error
func (foo Foo) Sum(argv Input,reply *int) error{
	println(argv.Input1,"+",argv.Input2,"=",argv.Input1+argv.Input2)
	*reply=argv.Input1 +argv.Input2
	return nil
}

func startServer(){
	// 1. 注册foo，作为一个service
	var foo Foo
	if err := geerpc.Register(&foo);err != nil {
		log.Fatal("register err",err)
	}
	// 2. 开启server监听
	l,err := net.Listen("tcp","localhost:8022")
	if err!=nil{
		log.Fatal("network error:",err)
	}
	log.Println("start rpc server on",l.Addr())
	geerpc.Accept(l)
}

func main() {
	if os.Getenv("mode")=="server"{
		startServer()
	}else{
		log.SetFlags(0)
		client, _ := geerpc.Dial("tcp", "127.0.0.1:8022")
		defer func() { _ = client.Close() }()

		time.Sleep(time.Second)
		// send request & receive response
		//var wg sync.WaitGroup
		for i := 0; i < 1; i++ {
			//wg.Add(1)
			//go func(i int) {
			//	defer wg.Done()
				args := Input{
					Input1: 1,
					Input2: 2,
				}
				var reply int
				if err := client.Call("Foo.Sum", args, &reply); err != nil {
					log.Fatal("call Foo.Sum error:", err)
				}
				fmt.Printf("%d+%d=%d",args.Input1,args.Input2,reply)
			//}(i)
		}
		//wg.Wait()
	}
}