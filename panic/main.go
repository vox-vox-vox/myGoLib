package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	//routineNoRecover()
	routineRecover()
}

/*
	goroutine 未加 recover，panic时整个进程崩溃
*/
func routineNoRecover()  {
	println("main")
	go func() {
		println("goroutineA")
		panic("make panic")
	}()
	time.Sleep(10)
}

/*
Go语言没有异常系统，其使用 panic 触发宕机类似于其他语言的抛出异常，recover 的宕机恢复机制就对应其他语言中的 try/catch 机制。

*/
func routineRecover()  {
	println("main")
	go func() {
		// 延迟处理的函数
		defer func() {
			// 发生宕机时，获取panic传递的上下文并打印
			err := recover()
			switch err.(type) {
			case runtime.Error: // 运行时错误
				fmt.Println("runtime error:", err)
			default: // 非运行时错误
				fmt.Println("error:", err)
			}
		}()

		println("goroutineA")
		panic("make panic")
		print("afterPanic")
	}()
	time.Sleep(10)
}