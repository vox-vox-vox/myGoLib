package chanBase

/*
channel 数据结构如下

type hchan struct {
	// 循环队列
	qcount   uint				channel元素个数
	dataqsiz uint				channel循环队列长度
	buf      unsafe.Pointer		channel缓冲区数据指针
	sendx    uint				channel发送操作处理到的位置
	recvx    uint				channel接受操作处理到的位置

	// 元素类型
	elemsize uint16				channel元素大小
	elemtype *_type				channel元素类型

	closed   uint32


	recvq    waitq
	sendq    waitq

	lock mutex
}
*/

func dataStructure() {
	chanA := make(chan int)
	chanA <- 0
}
