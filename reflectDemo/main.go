package main

import (
	"fmt"
	"reflect"
	"reflectDemo/person"
)

/*
type
	Method
value

*/

func main() {
	var hx person.HX
	hx.Age=23
	hx.Name="huxiao"

	//Type
	typee:=reflect.TypeOf(hx)
	println(typee.Name()) // type-name
	/*
	尽管type-name可以自定义，type的种类却是有限的，一共有26种：
	const (
		Invalid Kind = iota
		Bool
		Int
		Int8
		Int16
		Int32
		Int64
		Uint
		Uint8
		Uint16
		Uint32
		Uint64
		Uintptr
		Float32
		Float64
		Complex64
		Complex128
		Array
		Chan
		Func
		Interface
		Map
		Ptr
		Slice
		String
		Struct
		UnsafePointer
	)
	*/
	println(typee.Kind())// 返回25，说明是struct

	//Elem
	//println(typee.Elem().Name())

	//Value
	value1:=reflect.New(typee)// 新建一个reflect-value , 这是一个指针
	fmt.Println(value1.Interface().(*person.HX))
	println("value1 kind : ",value1.Kind())// 返回22，说明是指针。这是Value的Kind()，注意和type的Kind()做区分，虽然都是Kind()函数。

	value2:=reflect.ValueOf(hx)
	println("value2 kind : ",value2.Kind())// 返回25，说明是struct
	println("value2 kind from type : ",value2.Type().Kind()) // todo: value.type.kind 和 value.kind 有什么区别

	//Method
	//println(typee.Method(1))// 如何调用type对应的方法


}
