package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
)

func RunStringValueDemos(rdb *redis.Client){
	//StringToStringValue(rdb)
	//StructToStringValue(rdb)
	IntToStringValue(rdb)
}

// 插入简单字符串
// 这是一个官方例子
func StringToStringValue(rdb *redis.Client){
	err := rdb.Set("key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get("key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := rdb.Get("key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exist

	// 将key删除
	rdb.Del("key")
}

func IntToStringValue(rdb *redis.Client){
	err := rdb.Set("key", 10, 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get("key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := rdb.Get("key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	// Output: key value
	// key2 does not exist

	// 将key删除
	rdb.Del("key")
}


type Person struct {
	Name string `json:"name"`
	Age int `json:"age"`
	School string `json:"school"`
}
// 插入struct
func StructToStringValue(rdb *redis.Client){
	hx:=Person{
		Name:   "huxiao",
		Age:    23,
		School: "SEU",
	}
	// 如果直接set struct，会报错："redis: can't marshal main.Person (implement encoding.BinaryMarshaler)"
	// 要自己先序列化，再传入redis
	data, _ :=json.Marshal(hx)
	err := rdb.Set("key",data,0).Err()
	if err != nil {
		panic(err)
	}

	var hx2 Person
	val2, err := rdb.Get("key").Result()
	if err == redis.Nil {
		fmt.Println("key does not exist")
	} else if err != nil {
		panic(err)
	} else {
		if err := json.Unmarshal([]byte(val2),&hx2);err != nil {
			return
		}
		fmt.Println("key2", hx2)
	}
}
