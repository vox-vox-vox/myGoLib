package main

// 本函数用于测试map拷贝时，重新修改拷贝后的map，会不会影响原map的值
// 结论：map本身是指针，所以这样的拷贝是没有作用的，只是将map2重新指向了map1的地址而已，
func main() {
	map1:=make(map[string]string)
	map1["key1"]="value1"
	map1["key2"]="value2"
	map1["key3"]="value3"
	map2:=map1
	map2["key1"]="newValue1"

	for _,v:=range map1{
		println(v)
	}
	for _,v:=range map2{
		println(v)
	}
}
