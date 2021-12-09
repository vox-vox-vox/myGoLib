package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/trace"
)

//

func main() {
	f,err:=os.Create("trace.out")
	if err!=nil{
		panic(err)
	}
	runtime.Gosched()
	defer f.Close()
	err = trace.Start(f)
	if err!=nil{
		panic(err)
	}

	fmt.Println("hello GMP! ")

	trace.Stop()
}
