package main

import (
	"fmt"
	"testing"
)

func TestRedis(t *testing.T) {

	cli := NewRedisClient("10.128.5.73:6379", "tjj.31415")
	if cli != nil {
		fmt.Println("redis连接成功！")
	} else {
		fmt.Println("redis连接失败！")
	}

}
