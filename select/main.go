package main

import "time"

func main() {
	// choseFast()
	// choseFastWithDeadLock()
	// runAny()
	defaultAndFor()
}

/*
2个函数选择一个执行最快的进行输出，以下代码会出现死锁问题。如果 select 的两个 channel 都没有东西输入，将会一直阻塞。
*/
func choseFast() {
	chanA := make(chan string)
	chanB := make(chan string)
	var output string
	go func() {
		output := funA()
		if output != "" {
			chanA <- output
		}
	}()
	go func() {
		output := funB()
		if output != "" {
			chanB <- output
		}
	}()
	select {
	case output = <-chanA:
	case output = <-chanB:
	}
	println(output)
}

/*
下面这个函数是模拟上面函数出现死锁的情况
*/
func choseFastWithDeadLock() {
	chanA := make(chan string)
	chanB := make(chan string)
	var output string
	go func() {
		output := funA()
		println(output) // 模拟调用funA()函数失败时不向chanA写信息
	}()
	go func() {
		output := funB()
		println(output) //  模拟调用funB()函数失败时不向chanB写信息
	}()
	select {
	case output = <-chanA:
	case output = <-chanB:
	}
	println(output)
}

func funA() string {
	time.Sleep(10 * time.Second)
	return "A"
}

func funB() string {
	time.Sleep(30 * time.Second)
	return "B"
}

/*
如果同时满足多个case，随机选择一个输出
*/
func runAny() {
	ch := make(chan int)
	go func() {
		for range time.Tick(1 * time.Second) {
			ch <- 0
		}
	}()

	for {
		select {
		case <-ch:
			println("case1")
		case <-ch:
			println("case2")
		}
	}
}

/*
default+for 导致死循环
*/
func defaultAndFor() {
	chanA := make(chan int)
	chanB := make(chan int)
	go func() {
		time.Sleep(1)
		chanA <- 0
	}()
	go func() {
		time.Sleep(200 * time.Millisecond)
		chanB <- 1
	}()
	for {
		select {
		case <-chanA:
			println("chanA received")
		case <-chanB:
			println("chanB received")
		default:
			println("default")
		}
	}
}
