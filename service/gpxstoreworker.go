package service

import (
	"go.uber.org/zap"
	"time"
	"zhituBackend/common"
	"zhituBackend/db"
	"zhituBackend/model"
)

// 这里是一个处理gpx数据的协程函数，负责写入到数据库中
// 分为2个不同的管道，分别接收单个数据以及批量数据
func worker(chGpx chan *model.GpxData, chList chan *model.GpxDataArray, cache *db.RedisClient) {
	for {
		select {
		case gpx, ok := <-chGpx:
			if !ok {
				common.Logger.Info("gpx data chan is closed ")
				break
			}
			//fmt.Println(gpx.ToJsonString())
			_, _, err := cache.AddGpx(gpx)
			if err != nil {
				common.Logger.Error("save gpx data to redis err: ", zap.Error(err))
			}

		case list, ok := <-chList:
			if !ok {
				common.Logger.Info("gpx array chan is closed ")
				break
			}
			//fmt.Println(list.ToJsonString())
			_, _, err := cache.AddGpxDataArray(list)
			if err != nil {
				common.Logger.Error("save gpx array data to redis err: ", zap.Error(err))
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
	common.Logger.Info("thread of gpx store worker exit here! ")
}

// 启动持久化服务
// http的存储服务都使用此worker协程池来负责保存
func StartGpxStoreWorker(cache *db.RedisClient) (chgpx chan *model.GpxData, chanArray chan *model.GpxDataArray) {
	chGpx := make(chan *model.GpxData, common.Config.Server.QueLen)
	chanArray = make(chan *model.GpxDataArray, common.Config.Server.QueLen)

	for i := 0; i < common.Config.Server.Workers; i++ {
		go worker(chGpx, chanArray, cache)
	}

	return chGpx, chanArray
}
