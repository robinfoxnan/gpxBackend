package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var config *Config = nil
var redisCli *RedisClient = nil

func main() {

	//testUser()
	//InitCassandra()
	//FindInboxData()

	rand.Seed(time.Now().UnixNano())

	// 加载同目录下的config.yaml文件，
	config = LoadConfig()
	//fmt.Println(config)

	// 创建全局变量logger
	CreateLogger()

	// 创建redis连接池子
	redisCli = NewRedisClient(config.Redis.RedisHost, config.Redis.RedisPwd)
	if redisCli == nil {
		//fmt.Println("can't connnet to server redis")
		logger.Error("can't connnet to server redis")
		return
	}

	// 启动存储数据的worker
	// 这里使用的是server文件中的全局变量
	chGpxData, chanList = StartStoreWorker(redisCli)
	startHttpServer()

}

func testRandom() {
	for i := 0; i < 100; i++ {
		fmt.Println(GetRandom64())
	}
}
func testUser() {
	user := NewUserInfo()
	user.Nick = "robin"
	data := AnyToMap(user)
	fmt.Println(data)

	data1 := make(map[string]string)
	data1["name"] = "robin"
	data1["id"] = "1001"
	user1 := NewUserInfo()
	err := AnyFromMapString[UserInfo](data1, user1)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(user1)

}

func test1() {
	//cli.TestAddData()
	//param := QueryParam{Phone: "13810501031", Friend: "13810501031", DateStr: "20220905", TmStart: 0, TmEnd: 0}

	//cli.FindGpxTrack(&param)
	//

	lastGpx, err1 := redisCli.FindLastGpx("13810501031")
	if err1 == nil {
		str, _ := lastGpx.ToJsonString()
		fmt.Println("最后活动点", str)
	}

	//cli1 := NewRedisClient(conf.Redis.RedisHost, conf.Redis.RedisPwd)
	//cli1.FindLastGpx("13810501031")
	var wg sync.WaitGroup
	for i := 1; i < 4; i++ {
		wg.Add(1) //增加信号量
		go doGet(redisCli, &wg)
	}
	wg.Wait()
	fmt.Println("finished ")
	//time.Sleep(15 * time.Second)
}

func doGet(cli *RedisClient, wg *sync.WaitGroup) {
	for i := 0; i < 1000; i++ {
		cli.FindLastGpx("13810501031")
	}
	wg.Done()
}
