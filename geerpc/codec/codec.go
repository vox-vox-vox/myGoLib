package codec

import "io"

/*
	err = client.Call("Arith.Multiply", args, &reply)
*/

type Header struct {
	ServiceMethod string //ServiceMethod 是服务名和方法名，通常与 Go 语言中的结构体和方法相映射。
	Seq uint64			//Seq 是请求的序号，也可以认为是某个请求的 ID，用来区分不同的请求。
	Error string		//Error 是错误信息，客户端置为空，服务端如果如果发生错误，将错误信息置于 Error 中。
}

// Codec 对消息进行编解码的接口，实现了接口就说明具有 读/写/关闭 网络文件的能力
type Codec interface{
	io.Closer
	ReadHeader(*Header) error			// 从对应的流中读取Head信息
	ReadBody(interface{}) error			// 从对应的流中读取Body信息
	Write(*Header,interface{}) error	// 向对应的流中写入Head，Body信息
}

type NewCodecFunc func(io.ReadWriteCloser) Codec // Codec的构造函数

type Type string

const(
	GobType Type="application/gob"
	JsonType Type="application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc //与工厂模式不同的是，得到的是构造函数而非实例

func init()  {
	NewCodecFuncMap=make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType]=NewGobCodec
}