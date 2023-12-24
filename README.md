# gpxBackend
这是一个travelbook的后端服务。使用golang1.18开发。

开发环境为windows11。

使用redis存储用户和地理信息数据。

使用了mongoDB存储社区相关的帖子与评论等内容；如果不使用这部分功能，可以不架设；

配置文件为：config.yaml

服务默认使用http协议，使用7817端口；

```
cd src/server
go build
./server.exe
```



linux下运行：

```
screen -S gpx
chmod 777 ./server
./server
ctrl a + d
```

debug目录是所需要的文件。