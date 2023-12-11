# gpxBackend
这是一个travelbook的后端服务。使用golang1.18开发。

开发环境为windows11。

数据使用redis存储。

配置文件为：config.yaml

服务默认使用http协议，使用7817端口；

```
cd src/server
go build
./server.exe
```



linux下

```
screen -S gpx
chmod 777 ./server
./server
ctrl a + d
```

