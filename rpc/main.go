package main

import (
	"log"
	"os"
	"rpc/client"
	"rpc/server"
)

func main() {
	if os.Getenv("mode")=="server"{
		server.StartServerAndRegister()
	}else{
		cli,err:=client.GetClient()
		if err!=nil{
			panic(err)
		}
		var reply string
		err = cli.Call("HelloService.Hello", "hello", &reply)
		if err != nil {
			log.Fatal(err)
		}
		println("rpc call output is: ",reply)
	}
}