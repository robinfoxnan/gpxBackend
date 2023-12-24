package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"zhituBackend/common"
	"zhituBackend/common/middleware"
	"zhituBackend/controller"
	"zhituBackend/db"
)

func startServer() {

	// 初始化雪花算法
	common.InitSnow(1)
	// 加载配置

	err := common.InitConfig("config.yaml")
	if err != nil {
		common.Logger.Error("load config.yaml err: %v", zap.Error(err))
		return
	}

	// 初始化redis
	err = db.InitRedisDb(common.Config.Redis.RedisHost, common.Config.Redis.RedisPwd)
	if err != nil {
		common.Logger.Error("init redis err: %v", zap.Error(err))
		return
	}

	// 初始化mongo
	err = db.InitMongoClient(common.Config.MongoDb.MongoHost, common.Config.MongoDb.DbName)
	if err != nil {
		common.Logger.Error("init mongo err: %v", zap.Error(err))
		return
	}

	//gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	//router.Use(gin.Recovery())

	router.Use(middleware.GinLogger(common.Logger), middleware.GinRecovery(common.Logger, true))
	controller.InitHanlder(router)

	// 使用Let's Encrypt证书
	//m := autocert.Manager{
	//	Prompt: autocert.AcceptTOS,
	//	//HostPolicy: autocert.HostWhitelist("localhost"), // 你的域名
	//	HostPolicy: nil,
	//	Cache:      autocert.DirCache("certs"), // 证书缓存目录
	//}

	host := common.Config.Server.Host + ":" + strconv.Itoa(common.Config.Server.Port)
	common.Logger.Info("server is starting at:", zap.String("mode", common.Config.Server.Schema), zap.String("host", host))

	if common.Config.Server.Schema == "https" {

		// 启动HTTP/2服务器
		server := &http.Server{
			Addr:    host, // ":443"
			Handler: router,
			//TLSConfig: &tls.Config{
			//	GetCertificate: m.GetCertificate,
			//},
		}
		server.ListenAndServeTLS(common.Config.Server.CertFile, common.Config.Server.KeyFile)

	} else {
		// "0.0.0.0:80"
		router.Run(host)
	}

	common.Logger.Info("server stop here.")
}
func waitSignal() {
	// 创建一个通道用于接收信号
	sigChan := make(chan os.Signal, 1)

	// 捕获 Ctrl+C 和 Kill 信号
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞程序，等待信号
	sigReceived := <-sigChan
	str := fmt.Sprintf("Received signal: %v\n", sigReceived)
	common.Logger.Info("signal ready to exit", zap.String("info", str))
	controller.StopGpxChannal()
}
func main() {
	//testFilename()
	go waitSignal()
	startServer()
}
