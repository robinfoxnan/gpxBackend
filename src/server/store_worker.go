package main

import "fmt"

// 这里是一个处理gpx数据的协程函数，负责写入到数据库中
// 分为2个不同的管道，分别接收单个数据以及批量数据
func worker(chGpx chan *GpxData, chList chan *GpxDataArray, cache *RedisClient) {
	for {
		select {
		case gpx, ok := <-chGpx:
			if !ok {
				fmt.Println("chan is closed")
				break
			}
			//fmt.Println(gpx.ToJsonString())
			_, _, err := cache.AddGpx(gpx)
			if err != nil {
				//logger.Error(err.Error())
			}

		case list, ok := <-chList:
			if !ok {
				fmt.Println("chan is closed")
				break
			}
			//fmt.Println(list.ToJsonString())
			_, _, err := cache.AddGpxDataArray(list)
			if err != nil {
				logger.Error(err.Error())
			}
		default:

		}
	}
	fmt.Println("exit\n")
}

// 启动持久化服务
// http的存储服务都使用此worker协程池来负责保存
func StartStoreWorker(cache *RedisClient) (chgpx chan *GpxData, chanArray chan *GpxDataArray) {
	chGpx := make(chan *GpxData, config.Server.QueLen)
	chanArray = make(chan *GpxDataArray, config.Server.QueLen)

	for i := 0; i < config.Server.Workers; i++ {
		go worker(chGpx, chanArray, cache)
	}

	return chGpx, chanArray
}
