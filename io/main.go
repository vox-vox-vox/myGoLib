package main

import (
	"log"
	"os"
)

func main() {

	//测试write
	f, err := os.Create("./io/forWrite.txt")
	if err != nil {
		log.Println("open file error:", err)
		return
	}
	defer f.Close()
	data := []byte("114514 前辈")
	_, err = f.Write(data)

}
