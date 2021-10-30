package main

import "fmt"

func main() {
	// 本函数用于测试map拷贝时，重新修改拷贝后的map，会不会影响原map的值
	// 结论：map本身是指针，所以这样的拷贝是没有作用的，只是将map2重新指向了map1的地址而已，map只能浅拷贝
	// map深拷贝只能采用循环赋值的方式实现
	// mapShallowCopy()

	// 本函数用于map深拷贝
	// 注意：struct赋值是深拷贝
	//mapDeepCopy()

	// map更改value的坑
	mapForRange()
}

func mapShallowCopy() {
	map1 := make(map[string]string)
	map1["key1"] = "value1"
	map1["key2"] = "value2"
	map1["key3"] = "value3"
	map2 := map1
	map2["key1"] = "newValue1"

	for _, v := range map1 {
		println(v)
	}
	for _, v := range map2 {
		println(v)
	}
}

func mapDeepCopy() {
	type Person struct {
		Name string
		Age  int
	}
	map1 := make(map[string]Person)
	map1["key1"] = Person{Name: "hx", Age: 23}
	map1["key2"] = Person{Name: "zqy", Age: 23}
	map1["key3"] = Person{Name: "llz", Age: 20}

	map2 := make(map[string]Person)
	for k, v := range map1 {
		map2[k] = v
	}

	map2["key1"] = Person{
		Name: "hhxx",
		Age:  100,
	}
	for _, v := range map1 {
		fmt.Println(v)
	}
	for _, v := range map2 {
		fmt.Println(v)
	}
}

func mapForRange() {
	type Person struct {
		Name string
		Age  int
	}
	map1 := make(map[string]Person)
	map1["key1"] = Person{Name: "hx", Age: 23}
	map1["key2"] = Person{Name: "zqy", Age: 23}
	map1["key3"] = Person{Name: "llz", Age: 20}

	map2 := make(map[string]Person)
	for k, v := range map1 {
		map2[k] = v
	}

	// 直接操作v会改变map中的value吗？
	for _, v := range map2 {
		v.Name = "newName"
	}

	for _, v := range map1 {
		fmt.Println(v)
	}
	for _, v := range map2 {
		fmt.Println(v)
	}

	// 要怎样才能批量更改value呢
	for k, v := range map2 {
		v.Name = "newName"
		map2[k] = v
	}

	for _, v := range map1 {
		fmt.Println(v)
	}
	for _, v := range map2 {
		fmt.Println(v)
	}

}
