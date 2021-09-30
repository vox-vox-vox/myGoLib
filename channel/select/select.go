package main

import "time"

// 2个函数选择一个执行最快的进行输出
func main()  {
	chanA:=make(chan string)
	chanB:=make(chan string)
	var output string
	go func() {
		output:=funA()
		if output!=""{
			chanA<-output
		}
	}()
	go func() {
		output:=funB()
		if output!=""{
			chanB<-output
		}
	}()
	select {
	case output=<-chanA:
	case output=<-chanB:
	}
	println(output)
}

func funA()  string{
	time.Sleep(10*time.Second)
	return "A"
}


func funB()  string{
	time.Sleep(30*time.Second)
	return "B"
}