package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct{
	conn io.ReadWriteCloser//conn 是由构建函数传入，通常是通过 TCP 或者 Unix 建立 socket 时得到的链接实例，可以看做网络文件
	buf *bufio.Writer//buf 是为了防止阻塞而创建的带缓冲的 Writer，一般这么做能提升性能。写网络文件时会写到buf中，对buf进行flush才会将其写进网络文件中
	dec *gob.Decoder//可以看做writer,Decoder(fd).Decode(interface{})==Write([]byte)
	enc *gob.Encoder//可以看做reader,Encoder(fd).Encode(interface{})==Read([]byte)
}

var _ Codec =(*GobCodec)(nil) // 确保GobCodec实现了Codec的所有方法：将nil转为 *GobCodec类型，再转换为Codec接口，如果转换失败，说明GobCodec并没有实现Codec接口的所有方法

func NewGobCodec(conn io.ReadWriteCloser) Codec{
	buf := bufio.NewWriter(conn)// 相当于在conn之上加了一层buffer，本身也可以作为文件
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),// 从网络文件中直接读取
		enc:  gob.NewEncoder(buf),// 实际上写到了buf里，flush时才写入网络文件
	}
}

// ReadHeader 从网络文件中读取Header
func (c *GobCodec) ReadHeader(header *Header) error{
	// todo: Gob 解码之后的东西放在哪？怎么返回啊
	log.Println("reading Header ...")
	err:=c.dec.Decode(header)
	log.Println("read Header over")
	return err
}

// ReadBody 从网络文件中读取Body
func (c *GobCodec) ReadBody(body interface{}) error{
	log.Println("reading Body ...")
	err:=c.dec.Decode(body)
	log.Println("read Body over")
	return err
}
// Write 向网络文件中写入header 和 body
func (c *GobCodec) Write(header *Header,body interface{}) (err error){
	// todo:defer func 是干啥的
	defer func() {
		_ = c.buf.Flush()//将buf中的内容Flush到网络文件中
		if err != nil{
			_ = c.Close()
		}
	}()
	if err:= c.enc.Encode(header);err!=nil{
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err:= c.enc.Encode(body);err!=nil{
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	return nil
}

// Close 关闭一个网络文件
func(c *GobCodec) Close() error{
	return c.conn.Close()
}