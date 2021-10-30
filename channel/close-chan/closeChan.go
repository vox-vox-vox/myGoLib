package main

import (
	"fmt"
	"strconv"
	"time"
)

// 为了打印goroutine情况，采用Test文件形式

func do(taskCh chan int) {
	for {
		select {
		case t := <-taskCh:
			time.Sleep(time.Millisecond)
			fmt.Printf("task %d is done\n", t)
		}
	}
}

func insertToClosedChan() {
	taskCh := make(chan int, 10) // chan with buffer
	go do(taskCh)
	// insert
	for i := 0; i < 20; i++ {
		if i<10{
			println("insert number" +strconv.Itoa(i)+" into chan")
			taskCh <- i
		}else if i==10{
			close(taskCh)
		}else{
			taskCh <- i
		}
	}
	time.Sleep(time.Second)
}

func main() {
	//getFromClosedChan()

	insertToClosedChan() // panic: send on closed channel

	time.Sleep(2*time.Second)
}