package main

import (
	"fmt"
	"strings"
)

func main() {
	/*需求1 判断字符串是否包含  ]  */
	contain("[1,2,3]","]")
}

func contain(s string, pattern string){
	flag:=strings.Contains(s,pattern)
	if flag{
		fmt.Println("yes")
	}else{
		fmt.Println("no")
	}
}
