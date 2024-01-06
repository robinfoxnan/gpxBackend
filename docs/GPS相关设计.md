## 一、需求与功能分析

### 1.1 用户管理功能

注册、销毁用户、锁定；

登录，退出登录；

设置基本信息，验证真实身份绑定信息；

加关注、取消关注、对粉丝设置权限、对关注对象设置备注；

建立群、添加到群、群验证、群成员移除、群解散；

### 1.2 GPX数据管理

上报实时单点数据，上报一周内的离线轨迹数据；

查询好友数据；



### 1.3 聊天功能

一对一单聊：是2个用户之间的聊天，以及用于与系统账户，或者机器人之间的聊天；

- 消息支持回执：服务收到回执，服务已经处理的回执，对方收到回执，对方阅读回执；

- 消息支持无限期撤回；（此功能要求消息必须具有唯一的ID)



多对多的群聊或者频道：

群聊类似微信的群聊，区分管理员与普通成员；

频道只有有限的成员可以发送消息，其他的大多数用户为只读用户；

- 消息支持回执：服务收到回执，服务已经处理的回执；
- 消息支持无限期撤回；（此功能要求消息必须具有唯一的ID)





后端的数据与其他的系统不同，这里不使用关系数据库，使用redis存储用户信息，好友关系，群组信息以及GPX信息；IM信息使用SCYLLADB来存储；

### 1.3.1 读写扩散问题

1. 每个用户都有一个发队列、收队列、回执队列；发队列记录用户所有发出的消息，在转发前可以过滤；收队列用于排队所有收到的单发消息与系统的通知消息；回执队列用于保存所有发消息的回执情况；撤回动作也属于发送消息的一种，保存于发队列；
2. 一对一聊天，直接写自己的发记录，同时写到对方的收队列中；服务和用户都需要对一对一消息进行应答；
3. 群组与频道一样，采用读扩散，多个用户同时写到群组唯一消息队列中，按照服务收到的时间戳进行排序；同时向所有在线的用户广播
4. 





### 1.3.2 如何保证不丢

1）每个用户（多终端）分别有自己的发送编号（uint64），服务端在收到消息后会做一系列的应答，提示消息的处理状态：

A：服务端收到，已经写入了消息队列；

B：写入到后端数据库，同时写到了分发队列；



### 1.3.3 消息撤回

是在消息队列中标记该消息已经被撤回；

一对一消息：需要通知对端用户在收队列中

### 1.3.4 消息删除

仅仅在客户端本地删除消息，远端不会标记消息；



## 二、服务接口设计



### 2.1 用户管理

首先，用户管理是最基础的东西，不然无法分享信息

```go
// 第一阶段，主要是注册好友，添加好友等所有好友的相关操作
http.HandleFunc("/v1/user/login", loginHandlers)   //
http.HandleFunc("/v1/user/logout", logoutHandlers) //

http.HandleFunc("/v1/user/regist", registUserHandlers)                 // 注册用户，目前只支持匿名方式
http.HandleFunc("/v1/user/searchfriends", searchUserHandlers)          // 搜索用户
http.HandleFunc("/v1/user/addfriendreq", addfriendreqHandlers)         // 发送加好友请求，如果这里使用微博的关注模式，则不需要应答
http.HandleFunc("/v1/user/addfriendres", addfriendresHandlers)         // 应答好友请求，默认不需要应答，以后需要支持开启确认
http.HandleFunc("/v1/user/setbaseinfo", setbaseinfoHandlers)           // 设置个人基本信息
http.HandleFunc("/v1/user/setrealinfo", setrealinfoHandlers)           // 设置个人的认证信息，包括手机，邮箱等
http.HandleFunc("/v1/user/removefriend", removefriendHandlers)         // 删除
http.HandleFunc("/v1/user/blockfriend", blockfriendHandlers)           // 拉黑，可以认为是设置权限的一种
http.HandleFunc("/v1/user/permissionfriend", permissionfriendHandlers) // 设置好友权限
http.HandleFunc("/v1/user/listfriends", listfriendsHandlers)           // 获取好友列表
```

#### 2.1.1 登录

默认认为使用HTTPS加密，所以传输过程暂时不考虑自己加密了。

客户端首先检查本地配置文件，查看是否有用户名与 口令，如果有尝试登录，如果跳转到登录界面，让用户填写信息；

登录的参数有:

用户ID、密码、会话ID，设备信息；

登录成功后，返回用户ID，好友列表；客户端在配置中保存口令，下次自动登录使用；

方式0：使用之前的会话ID来恢复对话；

http://localhost:7817/v1/user/login?uid=1005&sid=6812630841045951575&type=0

```json
{
    "state": "ok",
    "session": {
        "sid": 6812630841045951575,
        "id": 1005,
        "tm": "2023/09/20 16:54:50",
        "key": ""
    }
}
```

SID错误会返回如下：

```json
{
    "state": "fail",
    "code": 303,
    "deitail":"should login first"
}
```
方式1：使用用户密码来登录；返回会话信息；



http://localhost:7817/v1/user/login?uid=1005&pwd=111111&type=1

```json
{
    "state": "ok",
    "session": {
        "sid": 6812630841045951575,
        "id": 1005,
        "tm": "2023/09/20 16:45:17",
        "key": "nil"
    }
}
```

客户端需要自己保存SID信息，必要时可以自己保存密码；

密码错误：

```json
{
    "state": "fail",
    "code": 302,
    "detail":“wrong pwd”
}
```

其他的type值会返回错误：

```json
{
    "state": "fail",
    "deitail":“not finish”
}
```

基本登录

当前默认使用用户名口令方式登录；注册时候提取了主机的硬件信息，所以这里可以使用硬件信息以及用户ID方式登录，不需要口令；如果设置了口令，则可以使用密码验证。

| 类型  | 字段      | 备注                                                         |
| ----- | --------- | ------------------------------------------------------------ |
| 地址  | /v1/login |                                                              |
| 参数1 | type      | 类型：<br/>1为默认的注册方式匿名，分配号码，设置口令；<br/>2是使用手机号注册，发送验证码，<br/>3是邮件方式，发送验证码； |
| 参数2 | pwd       | 用户自己设置的口令                                           |
| 参数3 | code      | 验证码                                                       |

登录后会返回SessionID，客户端在连接到服务端成功后，首先询问*SessionId*是否可用，如果没有则需要登录或者注册，如果有，则可以直接登录。

#### 2.1.2 退出

注销主要是退出登录状态；



#### 2.1.3 注册

功能定义：

注册新用户，在用户静态信息表中添加相关信息，成功后返回用户唯一ID；（userId为64位整数，格式化为8位）

| 类型  | 字段            | 备注                                                         |
| ----- | --------------- | ------------------------------------------------------------ |
| 地址  | /v1/user/regist | ?type=1&code=1234&pwd=123456&username=aaaa                   |
| 参数1 | type            | 类型：<br/>1为默认的注册方式匿名，分配号码，设置口令；<br/>2是使用手机号注册，发送验证码，<br/>3是邮件方式，发送验证码； |
| 参数2 | pwd             | 用户自己设置的口令                                           |
| 参数3 | code            | 验证码                                                       |
| 参数4 | param           | 硬件的基本信息组成的编码                                     |
| 参数5 | username        | 起一个用户名，也就是昵称                                     |

例如：

http://localhost:7817/v1/user/regist?type=1&pwd=111111&username=robin

返回的信息如下：

```json
{
    "state": "ok",
    "user": {
        "id": 1002,
        "phone": "0",
        "icon": "sys:icon/1.jpg",
        "name": "momo",
        "nick": "momo",
        "email": "momo@local.com",
        "pwd": "123456",
        "temp_pwd": "123456",
        "last_pt": "{}",
        "last_pt_tm": "{}",
        "region": "unknow",
        "ipv4": "0.0.0.0",
        "wxid": "0",
        "age": 1,
        "gender": 1,
        "online": false
    }
}
```

类型1）向服务端，申请一个新的号码，成功后返回号码，

类型2）手机号注册，服务端需要向手机号发送短息；用户再次提交验证码，服务端分配一个ID号，同时绑定手机信息；

类型3）使用邮箱注册，与手机号注册流程类似；

**所有的新用户都会添加到用户哈希表users中**；redis使用哈希表方便查找用户，但是不方便枚举所有的的用户，这需要后续的数据库支持；

| http code | body                            | 备注 |
| --------- | ------------------------------- | ---- |
| 200       | {"status:"ok", "userId":1001, } |      |
|           |                                 |      |
|           |                                 |      |

关于序号的使用， 有几种选择：

**基于 UUID**

**基于数据库主键自增**

**基于数据库多实例主键自增**

**基于类 Snowflake 算法**

**基于 Redis 生成办法**

**基于美团的 Leaf 方案（ID 段、双 Buffer、动态调整 Step）**

1）雪花算法：https://cloud.tencent.com/developer/article/1520916

2）美团的LEAF算法，这个类似ORALCE的SEQ，可以预先取若干个。https://cloud.tencent.com/developer/article/2269082

3）微信的解决方案：https://www.open-open.com/lib/view/open1465867321684.html

4）目前使用最简单的方法，使用redis存储当前最大的id；



这里采用**基于 Redis 生成办法**来实现。此方案依赖于使用REDIS集群。

设置全局变量

| 变量      | 类型   | 起始值 |      |
| --------- | ------ | ------ | ---- |
| userIdMax | string | 1000   |      |

默认的用户从1000开始，

每次申请ID，则需要从REDIS中自增1，获取下一次的ID。

防止攻击：匿名方式无法防止对注册接口的攻击。只能添加验证码以及工作量证明之类的方式缓解。但是后期可以要求绑定邮箱或者手机号的方式清除。



#### 2.1.4 搜索用户

搜索用户在第1阶段主要是用REDIS中查找，使用ID精确搜索，那么ID大于userIDMax肯定是错误的，如果是在缓存中存在，那么应该存在一个u+ID的键，能找到则认为是存在，否则就是不存在的；
| http code | body                          | 备注 |
| --------- | ----------------------------- | ---- |
| 200       | {"status:"ok", "user": "", }  |      |
| 200       | {"status:"fail", "user": "",} |      |

例如：http://localhost:7817/v1/user/searchfriends?fid=1005

```json
{
    "state": "ok",
    "user": {
        "id": 1005,
        "phone": "0",
        "icon": "sys:icon/1.jpg",
        "name": "robin",
        "nick": "robin",
        "email": "momo@local.com",
        "pwd": "",
        "temppwd": "",
        "region": "unknow",
        "ipv4": "0.0.0.0",
        "wxid": "0",
        "age": 1,
        "gender": 1,
        "tm": "2023/09/20 16:04:18",
        "sid": 0
    }
}
```

返回的数据，如果user的id不对，说明没有找到

#### 2.1.5 添加好友（关注）与删除、显示、不显示位置

正常情况下添加好友需要对方确认，这里为了简化设计，我们先采取关注的方式，也就是直接将对方添加到我们的关注列表，user following，哈希类型 ，键名使用uf+ID;

键值对分别使用对方的ID以及值使用JSON编码，包括对方的昵称，头像，对方是否可以查看自己的位置信息等权限。默认是向对方开放所有权限的；这里隐含的逻辑是查看对方则必须要先开放自己的信息，对等原则；

```
http://localhost:7817/v1/user/addfriendreq?uid=1005&sid=6812630841045951575&fid=1004
```

在目前的简单模式下，分别查找并关注对方就是好友了；

成功后，返回

```json
{
    "state": "ok",
    "detail": "add friend ok"
}
```

同时看到redis中设置了

用户uf1005的关注列表中添加了1004字段

```json
{
    "id": 1004,
    "icon": "sys:icon/1.jpg",
    "ctm": 1695200220490,
    "alias": "robin",
    "label": "",
    "visible": 7,
    "block": false
}
```

用户粉丝列表中，设置了权限掩码：us1004

```
ffff
```

更改好友是否显示的属性以及删除好友

```
http://localhost:7817/v1/user/setfriendinfo?uid=1005&fid=1008&sid=6812630841045951575&param=show
http://localhost:7817/v1/user/setfriendinfo?uid=1005&sid=6812630841045951575&param=ignore
http://localhost:7817/v1/user/setfriendinfo?uid=1005&sid=6812630841045951575&param=remove
```





#### 2.1.6 好友确认

在目前阶段，不需要确认，只是需要双向关注即可，类似小红书与微博；虽然目前没有计划制作动态图文等功能；



#### 2.1.7 个人基本信息

登录状态随时可以更改，这里其实最核心的主要是头像，方便对方查看，昵称不可以随便更改，防止假冒；

http://localhost:7817/v1/user/setbaseinfo?

GET命令不支持

```json
{
    "state": "fail",
    "deitail": "not finish"
}
```

POST命令，发送body为

```json
{
        "id": 1005,
        "phone": "",
        "icon": "sys:icon/2.jpg",
        "name": "robin",
        "nick": "飞鸟",
        "email": "",
        "pwd": "123456",
        "temppwd": "",
        "region": "unknow",
        "ipv4": "0.0.0.0",
        "wxid": "0",
        "age": 42,
        "gender": 1,
        "tm": "2023/09/20 16:04:18",
        "sid": 6812630841045951575
    }
```

成功后返回：

```json
{
    "state": "ok",
    "deitail": "user json updated"
}
```



#### 2.1.8 个人认证信息

这个在后续的阶段实现；

#### 2.1.9 删除

取消关注

#### 2.1.10 拉黑

这个需要设置标志；



#### 2.1.11 更改向粉丝开放权限

更改向对方

http://localhost:7817/v1/user/permissionfriend?

参数：

```json
{
    "id": 1005,
    "fid":1004,
    "sid":6812630841045951575,
    "opt": "-a"
}
```

-apt   +apt

结果

```json
{
    "state": "ok",
    "deitail": "friend permission updated"
}
```

此时通过获取粉丝列表可以查看粉丝的权限变化，掩码值有变化

#### 2.1.12 获取关注以及粉丝列表

关注列表中，需要体现是双向关注或者是单向关注。

http://localhost:7817/v1/user/listfriends?type=1&uid=1005&sid=6812630841045951575

```json
{
    "state": "ok",
    "count": 1,
    "list": [
        {
            "id": 1004,
            "icon": "sys:icon/1.jpg",
            "ctm": 1695200220490,
            "alias": "robin",
            "label": "",
            "visible": 7,
            "block": false
        }
    ]
}
```

http://localhost:7817/v1/user/listfriends?type=2&uid=1005&sid=6812630841045951575

```json
{
    "state": "ok",
    "count": 1,
    "list": [
        {
            "id": 1004,
            "mask": 65535
        }
    ]
}
```

这里返回的列表不同，主要是因为考虑到粉丝可能会非常的多，而关注的不多，所以粉丝部分仅仅保存了权限掩码，需要客户端二次获取粉丝信息补全；

后续会完善相关的部分；


### 3. 数据上报

#### 3.1 单点上报

上报的数据格式为：

http://localhost:7817/v1/gpx/updatepoint

```json
 {
       "uid":"1008",
     "lat":40.00651,
     "lon":116.258824,
     "ele":0,
     "accuracy":20,
     "src":"wifi",
     "direction":34.578010847945244,
     "city":"北京市",
     "addr":"北京市海淀区香山路",
     "street":"香山路",
     "streetNo":"",
     "speed":0,
     "tm":1699950389,
     "tmStr":"23-11-318-04-26-29"
 }
```

成功为：

```
{"state": "OK" }
```

数据会存储在键：ugpx1005_19700101中

同时添加用户临时状态数据，up1005键

#### 3.2 单点位置查询

http://localhost:7817/v1/gpx/position?

```json
{
    "uid":"1005",
    "sid":"6812630841045951575",
    "fid":"1004"
}
```

这里会检查1005会话状态，并检查1005在1004粉丝列表中的权限，

返回数据为：

```json
{
    "state": "OK",
    "pt": {
        "uid": "1004",
        "lat": 40,
        "lon": 116,
        "ele": 100,
        "speed": 0,
        "tm": 1007
    }
}
```

#### 3.3 上传批量的历史数据

```
http://localhost:7817/v1/gpx/v1/gpx/updatepoints
```

post 的数据示例为：

```json
[
	 {
	"uid": "1005",
	"lat" : 40,
	"lon":116,
	"ele" :100,
	"speed" :0,
	"tm":10000
	},
     {
	"uid": "1005",
	"lat" : 40,
	"lon":116,
	"ele" :100,
	"speed" :0,
	"tm":10001
	}
]
```



### 4 上报数据

功能：先查询当日已经上报数据的时间点，并上报剩余的点；服务端返回争取存储的最后时间点；

如果多日未上线，则上报7天内的记录数据；数据编码使用JSON编码

这里列举gpx格式中点的部分如下：

```
<trkpt lat="40.00533616" lon="116.25281886">
<ele>58.2550048828125</ele>
<time>2022-08-29T23:46:44.000Z</time>
<speed>0.0</speed>
<geoidheight>-8.0</geoidheight>
<src>gps</src>
<sat>3</sat>
<hdop>1.0</hdop>
<vdop>0.8</vdop>
<pdop>1.3</pdop>
</trkpt>
```

 #### 4.1查询位置

功能：本人位置信息通过GPS获得，主要是查询群组中所有人位置，以及查询某个好友位置；

查询好友需要检查好友的设置，是否屏蔽（是否需要这个功能）









查询轨迹

返回的信息为了简化，不使用json，而使用固定的格式：

```
tm64|40.00533616|116.25281886|58.2550048828125|0.0
```

### 5. 消息管理

#### 5.1 系统消息

个人基础信息绑定（头像图片使用链接）：

好友权限设置；

个人信息访问权限设置；

实名认证以及密码重置；

好友请求：请求添加好友

好友应答：同意，拒绝；

删除好友；

拉黑好友；

请求加入组，

邀请加入组

组权限设置：

组动作通知：

退出组

踢出组

解散组：



系统广播



#### 5.2 P2P消息

这里使用scyllaDb作为时序数据库，存储消息；

每个用户都有用ID命名的2个队列，单发队列以及多发队列；

所有一对一的用户消息，以及系统发送的广播，会写入用户的收队列；

用户发送的是一对一，系统消息，也写入到该用户的单发送队列中（因为如果切换了终端，这些消息可以认为是收的消息）；

也就是发送消息时候，需要写到2个队列，比如A发消息给B，同时在A和B的队列中写入数据；

群组发言作为备份写到用户多发队列中；

- 这里可以看到：一对一单聊，可以认为是写扩散的，直接写到收用户的收件箱中；自己发的消息是为了排序方便，也同时是为了多终端协作；
- 多发的队列，以后也许可以直接取消；目前主要是考虑容错；（暂不实现）



注意：这里与TINODE的会话为中心的方式不同，以会话为中心的方式，需要在服务端维护用户的会话列表；

而且每次需要拉去会话的消息，需要一个进入会话的动作；它这种做法造成一个BUG，删除当前会话，就无法通信了，等于把对方拉到黑名单了，不知道这么设计的还是BUG，但是不符合我们的习惯；

这里不采用会话作为存储逻辑主要是原因是无法知道有多少个会话在进行；而是在客户端自行本地存储会话列表；客户端自己去服务端拉数据的时候就可以根据时间点拉数据



文本消息

图片消息

视频消息

链接消息

位置信息

#### 5.3 P2G消息

多对多的消息针对每个聊天组都建立一个消息队列，所有人的消息都存储于该队列，如果需要撤回，则只需要标记一下即可；多人写时，通过redis计数器来做分布式锁，来控制同步问题，同时也有一个编号。

这个是读扩展的一个策略，后续如果用户多了需要扩容，首先考虑使用redis集群加速。





#### 5.4 文件消息

包括图片，音频，视频等各种使用文件服务方式上传和下载的；







### 6. 消息存储技术

所有消息支持3种确认：

当消息写入数据库时，有个发送完成的确认：1；在有kafka消息队列的时候，写入队列就认为完成了；没有消息队列时候使用服务内存缓冲一下，然后写入scyllaDB完成；

当消息发送到对方是，有个送达的确认:2；

当消息被对方读以后，有个已读的确认：4；





### 7. 文件上传与下载

#### 7.1 上传

http://127.0.0.1/uploadfile
https://8.140.203.92:7817/uploadfile

post 方式

参数有2个：

```
"file" 是file类型字段，用于传递文件，这里不分片，只接收10M以下的文件；
"src" 是用于表示客户端来源，如果是webpage，用于测试，返回HTML页面，否则返回JSON
```

返回的：

```json
{
    "des": "File uploaded successfully",
    "filename": "填表示例.jpg",
    "newname": "8767235711196729344.jpg",
    "state": "OK"
}
```





#### 7.2 下载

http://127.0.01/file/8767235711196729344.jpg?sid=123456

路径的后面部分是上传得到的路径，也就是客户端需要自己拼接路径，并带上Sid

### 8. 社区分享

格式定义： 包含文字内容与图片，

```json
{
	"nid": "",
    "uid": "1001",
    "nick": "飞鸟",
    "icon": "sys:9",
    "lat": 40.0,
    "log":116.0,
    "alt":0.0,
     "tm":0,
    "title": "颐和园雪后",
    "content":"颐和园雪后是一片冰封，万里雪飘",
    "images":["8767235711196729344.jpg"],
    "tags":["颐和园", "桂花"],
    "type": "point",  // track
	"trackfile": "8767235711196729344.json",
	"likes":0,
	"favs":0
}
```



#### 8.1 发布帖子

每个帖子都新分配一个号码，从"news_id"中自增获取；

POST http://localhost:7817/v1/news/publish

返回的

```JSON
{
    "code": 0,
    "des": "save news ok",
    "nid": "1001",
    "state": "ok"
}
```

#### 8.2 查询最近的新闻

GET  http://localhost:7817/v1/news/recent?page=1&size=20



```json
{
    "code": 0,
    "data": [
        {
            "nid": "1004",
            "uid": "1001",
            "nick": "飞鸟",
            "icon": "sys:9",
            "lat": 40,
            "log": 116,
            "alt": 0,
            "tm": 1703129301982,
            "title": "故宫纯色",
            "content": "颐和园雪后是一片冰封，万里雪飘",
            "images": [
                "8767235711196729344.jpg"
            ],
            "tags": [
                "颐和园",
                "桂花"
            ],
            "type": "point",
            "trackfile": "8767235711196729344.json",
            "likes": 0,
            "favs": 0
        },
        {
            "nid": "1003",
            "uid": "1001",
            "nick": "飞鸟",
            "icon": "sys:9",
            "lat": 40,
            "log": 116,
            "alt": 0,
            "tm": 1703128844826,
            "title": "颐和园雪后",
            "content": "颐和园雪后是一片冰封，万里雪飘",
            "images": [
                "8767235711196729344.jpg"
            ],
            "tags": [
                "颐和园",
                "桂花"
            ],
            "type": "point",
            "trackfile": "8767235711196729344.json",
            "likes": 0,
            "favs": 0
        }
    ],
    "des": "save news ok",
    "state": "ok"
}
```



#### 8.3 根据地点查询

GET http://localhost:7817/v1/news/byloc?uid=1008&sid=4553374183501693351&lat=40.0&lon=116.0&radius=5.0

```json
{
    "code": 0,
    "data": [
        {
            "nid": "1014",
            "uid": "1008",
            "nick": "飞鸟真人",
            "icon": "sys:9",
            "lat": 40,
            "log": 116,
            "alt": 0,
            "tm": 1703391463027,
            "title": "圆明园",
            "content": "颐和园雪后是一片冰封，万里雪飘",
            "images": [
                "8767235711196729344.jpg"
            ],
            "tags": [
                "颐和园",
                "桂花"
            ],
            "type": "point",
            "trackfile": "8767235711196729344.json",
            "likes": 0,
            "favs": 0,
            "deleted": false,
            "deltm": 0
        },
        {
            "nid": "1015",
            "uid": "1008",
            "nick": "飞鸟真人",
            "icon": "sys:9",
            "lat": 40,
            "log": 116,
            "alt": 0,
            "tm": 1703391463973,
            "title": "圆明园",
            "content": "颐和园雪后是一片冰封，万里雪飘",
            "images": [
                "8767235711196729344.jpg"
            ],
            "tags": [
                "颐和园",
                "桂花"
            ],
            "type": "point",
            "trackfile": "8767235711196729344.json",
            "likes": 0,
            "favs": 0,
            "deleted": false,
            "deltm": 0
        }
    ],
    "des": "find news ok",
    "state": "ok"
}
```



#### 8.4 关键字查询

http://localhost:7817/v1/news/bytag?tag=颐和

```

```

#### 8.5 删除新闻

```
http://localhost:7817/v1/news/delete?nid=1010&uid=1008&sid=4553374183501693351
```



#### 8.5 添加评论

POST http://localhost:7817/v1/news/addcomment?nid=1010&uid=1008&sid=4553374183501693351

```json
{
	"nid": "1010",
	"cid": "",
    "uid": "1008",
    "nick": "飞鸟",
    "icon": "sys:9",

     "tm":0,
     "pnid": "0",
     "toid": "1002",
     "tonick": "天涯剑客",

    "content":"颐和园雪后是一片冰封，万里雪飘",
    "images":["8767235711196729344.jpg"],
	"likes":0
}
```



#### 8.6 删除评论

http://localhost:7817/v1/news/deletecomment?nid=1010&uid=1008&cid=1005&sid=4553374183501693351

```json
{
    "code": 0,
    "des": "delete comment ok",
    "nid": "1010",
    "state": "ok"
}
```



#### 8.7 查询评论

http://localhost:7817/v1/news/findcomment?nid=1010&uid=1008&sid=4553374183501693351&page=1&size=20



```json
{
    "code": 0,
    "data": [
        {
            "nid": "1010",
            "cid": "1007",
            "uid": "1008",
            "nick": "飞鸟",
            "icon": "sys:9",
            "pnid": "0",
            "tm": 1703390186357,
            "toid": "1002",
            "tonick": "天涯剑客",
            "content": "颐和园雪后是一片冰封，万里雪飘",
            "images": [
                "8767235711196729344.jpg"
            ],
            "likes": 0,
            "deleted": false
        },
        {
            "nid": "1010",
            "cid": "1006",
            "uid": "1008",
            "nick": "飞鸟",
            "icon": "sys:9",
            "pnid": "0",
            "tm": 1703390185578,
            "toid": "1002",
            "tonick": "天涯剑客",
            "content": "颐和园雪后是一片冰封，万里雪飘",
            "images": [
                "8767235711196729344.jpg"
            ],
            "likes": 0,
            "deleted": false
        }
    ],
    "des": "find comment ok",
    "state": "ok"
}
```



#### 8.8 反馈报告

http://localhost:7817/v1/news/report?nid=1010&uid=1008&sid=4553374183501693351&

```
{
	"nid": "",
    "uid": "1008",
    "nick": "飞鸟真人",
    "icon": "sys:9",
    "lat": 40.0,
    "log":116.0,
    "alt":0.0,
     "tm":0,
    "title": "圆明园",
    "content":"有点问题",
    "images":[],
    "tags":[],
    "type": "report",  // track
	"trackfile": "8767235711196729344.json",
	"likes":0,
	"favs":0
}
```



```
{
    "code": 0,
    "des": "i know, and will sort it later",
    "nid": "1011",
    "state": "ok"
}
```



## 三、缓存设计

redis中的各个字段设置：

### 1.  用户ID的计数表

### 2. users_id

  普通字符串，从1000开始给用户分配ID号；

用户注册

### 3. 用户基础信息表
users，哈希表

key是用户的ID，比如管理员为u1000，值是json编码的基本信息，用户更新时候直接更新值；查找用户时候也可以提取信息，与每个用户单独存储相比，这样的设计可以防止命名空间过于混乱；

如果每个用户单独使用一个键来存储，更新会更加方便；
| key     | value |        |
| ------- | ----- | ------ |
| name    |       |        |
| age     |       |        |
| area    |       |        |
| gender  |       |        |
| phone   |       |        |
| email   |       |        |




### 4. 用户动态信息表
uinfo+ID，哈希表

 每个用户使用一个哈希键来存储，里面包括密码状态，当前位置等信息；

比如u1000，

| key     | value |        |
| ------- | ----- | ------ |
| name    |       |        |
| pt      |       |        |
| logintm |       | 时间戳 |
| ptm     |       |        |
| state   |       |        |

搜索一个人的信息就可以从这里获取；



###  5. 好友关系表管理
这里的好友设计类似社区的方式，使用双向关注的方式处理，主要是节省了好友验证的一步

也就是每个用户都一个关注列表和粉丝列表，通过控制粉丝的权限来决定向对方开放的权限；



uf+ID 使用哈希表存储
当用户添加了一个关注，则需要添加一个一个key，值需要设置对关注对象设置的属性，使用JSON编码；

| 属性    | 类型   | 说明             |
| ------- | ------ | ---------------- |
| alias   |        | 自定名标签       |
| label   | string | 标签，分组的名字 |
| visible | int    | 是否查看对方信息 |
| block   | bool   | 是否拉黑对方     |
|         |        |                  |
|         |        |                  |

粉丝列表

us + ID 哈希表，值仅仅存储一个粉丝的对应的权限；



### 6. gpx位置信息

程序见gpxdata.go

用户更新了自己的位置信息同时保存在2个地方，一个是用户的动态信息表中，一个是用户当天的轨迹队列中；

所以当请求好友信息的时候需要查询多个键值，这样效率不是很高，后续支持群组，将所有用户动态信息都放置在一个组中则可以提高查询效率；

查询用户的位置有2类参数：分别用于查询当前点位置，以及轨迹信息

```json
type QueryParamTrack struct {
	Uid     string `json:"uid" `
	Sid     string `json:"sid"`
	Fid     string `json:"fid" `
	DateStr string `json:"date" `
	TmStart int64  `json:"tmStart" `
	TmEnd   int64  `json:"tmEnd" `
}

type QueryParamPoint struct {
	Uid string `json:"uid" `
	Sid string `json:"sid"`
	fid string `json:"fid" `
}
```

存储点信息的数据结构为

```json
type GpxData struct {
	Uid   string  `json:"uid" `
	Lat   float64 `json:"lat" `
	Lon   float64 `json:"lon" `
	Ele   float64 `json:"ele" `
	Speed float64 `json:"speed" `
	Tm    int64   `json:"tm" ` // second time stamp
	//	Rest  int8    `json: "rest" `  // 1：处于静止中， 0，未设置
}
```

查找用户当前位置信息就保存在 up+Id的哈希表中，用pt字段存储JSON来表示位置；

#### 6.1 添加单点数据

首先查看动态表时间戳是否是旧数据，如果不需要更新则跳过；

更新动态表的当前位置信息，使用JSON编码；

用户的轨迹数据使用ugpx+id+"_"+ 日期，使用左侧推进队列的方式，保存7天；

在插入新的点前，如果发现是太近了，则认为在静止，则不进行插入队列；

如果有一定的距离，左侧推入队列；

#### 6.2 添加一组数据

先排序，之后检查最后一个点是否是否也是陈旧的数据；如果全都是旧数据直接跳过；

如果超过了7天，也不再保存了，意义不大了；

否则，使用最后一个点更新当前的位置信息；之后诸个点按照日期添加到不同日期的队列中；

最后对各个日期的队列设置超时时长为7天；



#### 6.3 查找单点实时位置

这个比较简单，先查看一下权限，就是uid在对方的粉丝列表中是否允许查看，

然后直接从FindLastGpx  



#### 6.4 查找组成员实时位置

暂无

#### 6.5 查询轨迹信息

```go
type QueryParamTrack struct {
	Uid     string `json:"uid" `
	Sid     string `json:"sid"`
	Fid     string `json:"fid" `
	DateStr string `json:"date" `
	TmStart int64  `json:"tmStart" `
	TmEnd   int64  `json:"tmEnd" `
}
```

目前轨迹信息只支持某一天的数据查询，也就是日期是必须要设置的，之后时间起始点和结束点应该在该日期的时间的范围内，否则就不直接返回不过滤时间段；

由于对不动的点进行了过滤，所以，为了防止过滤范围内没有数据，则应该尝试找到前一个点的位置；





四 安卓中需要注意的

1）使用http明文，需要开启权限

<uses-permission android:name="android.permission.INTERNET" />

```
<application
      
        android:usesCleartextTraffic="true"
```

2) 主线程中使用网络访问，

方法一：在activity的oncreate中添加 super前面添加，

```
 StrictMode.setThreadPolicy(
            ThreadPolicy.Builder()
                .detectDiskReads().detectDiskWrites().detectNetwork()
                .penaltyLog().build()
        )
        StrictMode.setVmPolicy(
            VmPolicy.Builder()
                .detectLeakedSqlLiteObjects().detectLeakedClosableObjects()
                .penaltyLog().penaltyDeath().build()
        )
```

但是不稳定，

或者使用异步方式，在loginviewmodel中

```kotlin
 fun login(username: String, password: String, server : HttpService?) {
        // can be launched in a separate asynchronous job

        object : Thread() {
            override fun run() {
                //网络操作连接的代码
                val result = loginRepository.login(username, password, server)

                if (result is Result.Success) {
                    _loginResult.value =
                        LoginResult(success = LoggedInUserView(displayName = result.data.displayName))
                } else {
                    _loginResult.value = LoginResult(error = R.string.login_failed)
                }
            }
        }.start()

    }
```



```
http://127.0.0.1:7817/v1/user/login?uid=1005&pwd=123456&type=1

http://10.0.2.2:7817/v1/user/login?uid=1005&pwd=123456&type=1
```

```
val fakeUser = LoggedInUser("", "", "", "0", "")
//构建url地址
var url1 = "${schema}://${host}/v1/user/login?type=1&pwd=${pwd}&uid=${uid}"
val fuel = FuelBuilder().build()

val res = Fuel.get(url1)
val responseData = res.body.toString()

val jsonObject = JSONObject(responseData)
val state = jsonObject.getString("state")
//LogHelper.d(" upload date response $state")
if (state == "ok") {
    val user = jsonObject.getJSONObject("session")
    if (user != null) {
        fakeUser.id = user.getLong("id").toString()
        fakeUser.sid = user.getLong("sid").toString()
    }
}
```





## 四、技术点

```
go run D:\Program\Go\src\crypto\tls\generate_cert.go -h

go run D:\Program\Go\src\crypto\tls\generate_cert.go -host bird2fish.com

```

### 4.1轨迹点简化算法

当需要精简经纬度轨迹点时，通常的方法是应用轨迹压缩算法，这有助于减少点的数量，同时保留轨迹的基本形状。有许多不同的轨迹压缩算法，其中一种常见的是道格拉斯-普克（Douglas-Peucker）算法。该算法基于递归，可以在一定误差范围内简化轨迹。

以下是一个简化版本的Douglas-Peucker算法的示例实现，你可以根据实际需要进行修改和扩展：

```go
package main

import (
	"fmt"
	"math"
)

type Point struct {
	Latitude  float64
	Longitude float64
}

func simplifyDouglasPeucker(points []Point, tolerance float64) []Point {
	if len(points) <= 2 {
		return points
	}

	// 找到距离最大的点
	maxDistance := 0.0
	maxIndex := 0

	for i := 1; i < len(points)-1; i++ {
		distance := distanceToSegment(points[i], points[0], points[len(points)-1])
		if distance > maxDistance {
			maxDistance = distance
			maxIndex = i
		}
	}

	// 如果最大距离小于阈值，直接返回首尾点
	if maxDistance <= tolerance {
		return []Point{points[0], points[len(points)-1]}
	}

	// 递归简化
	leftPart := simplifyDouglasPeucker(points[:maxIndex+1], tolerance)
	rightPart := simplifyDouglasPeucker(points[maxIndex:], tolerance)

	return append(leftPart[:len(leftPart)-1], rightPart...)
}

func distanceToSegment(p, start, end Point) float64 {
	lineLength := distanceBetweenPoints(start, end)

	if lineLength == 0 {
		return distanceBetweenPoints(p, start)
	}

	u := ((p.Latitude-start.Latitude)*(end.Latitude-start.Latitude) +
		(p.Longitude-start.Longitude)*(end.Longitude-start.Longitude)) / (lineLength * lineLength)

	if u < 0 || u > 1 {
		closestPoint := start
		if u > 0 {
			closestPoint = end
		}
		return distanceBetweenPoints(p, closestPoint)
	}

	intersection := Point{
		Latitude:  start.Latitude + u*(end.Latitude-start.Latitude),
		Longitude: start.Longitude + u*(end.Longitude-start.Longitude),
	}

	return distanceBetweenPoints(p, intersection)
}

func distanceBetweenPoints(p1, p2 Point) float64 {
	// 使用简化的球面距离计算
	radius := 6371.0 // 地球半径，单位：千米
	dLat := degToRad(p2.Latitude - p1.Latitude)
	dLon := degToRad(p2.Longitude - p1.Longitude)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(degToRad(p1.Latitude))*math.Cos(degToRad(p2.Latitude))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := radius * c
	return distance
}

func degToRad(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func main() {
	// 示例经纬度轨迹点
	trajectory := []Point{
		{Latitude: 38.115556, Longitude: 13.361389},
		{Latitude: 37.502669, Longitude: 15.087269},
		{Latitude: 38.115556, Longitude: 13.361389},
		{Latitude: 37.502669, Longitude: 15.087269},
		// 添加更多轨迹点...
	}

	// 误差阈值，根据实际情况调整
	tolerance := 1.0

	// 应用Douglas-Peucker算法进行轨迹简化
	simplifiedTrajectory := simplifyDouglasPeucker(trajectory, tolerance)

	// 输出简化后的轨迹点
	fmt.Println("简化后的轨迹点：", simplifiedTrajectory)
}

```

在这个例子中，`simplifyDouglasPeucker` 函数是Douglas-Peucker算法的实现，根据给定的误差阈值 `tolerance` 对轨迹进行简化。这个算法会递归地将轨迹简化为一系列新的轨迹点。根据你的需求，你可能需要调整阈值以及其他参数。这个实现是一个简单的示例，具体的应用场景可能需要更多的定制。



Douglas-Peucker算法的一个特性是，它能够保留原始轨迹中的主要转向位置（拐点）。由于算法的工作方式，当误差阈值 `tolerance` 设置得适当时，将保留原始轨迹中的重要转向位置，同时删除不太显著的直线部分。

但是，要注意的是，Douglas-Peucker算法不能保证精确地选择所有转向位置。选择的点数和具体的保留点取决于误差阈值和轨迹的几何形状。调整 `tolerance` 可以影响简化后轨迹的质量，但是在某些情况下，可能会丢失一些小的转向位置，特别是在轨迹段之间的变化较小的地方。

如果你的应用场景对于保留所有转向位置非常敏感，可能需要考虑其他的轨迹压缩算法，或者通过调整 `tolerance` 和其他参数来优化Douglas-Peucker算法以满足你的需求。在实际应用中，通常需要进行多次测试和调整，以找到最适合你数据集和应用场景的参数。



```kotlin
import kotlin.math.pow
import kotlin.math.sqrt

data class Point(val latitude: Double, val longitude: Double)

fun simplifyDouglasPeucker(points: List<Point>, tolerance: Double): List<Point> {
    if (points.size <= 2) {
        return points
    }

    // 找到距离最大的点
    var maxDistance = 0.0
    var maxIndex = 0

    for (i in 1 until points.size - 1) {
        val distance = distanceToSegment(points[i], points[0], points[points.size - 1])
        if (distance > maxDistance) {
            maxDistance = distance
            maxIndex = i
        }
    }

    // 如果最大距离小于阈值，直接返回首尾点
    if (maxDistance <= tolerance) {
        return listOf(points[0], points[points.size - 1])
    }

    // 递归简化
    val leftPart = simplifyDouglasPeucker(points.subList(0, maxIndex + 1), tolerance)
    val rightPart = simplifyDouglasPeucker(points.subList(maxIndex, points.size), tolerance)

    return leftPart.subList(0, leftPart.size - 1) + rightPart
}

fun distanceToSegment(p: Point, start: Point, end: Point): Double {
    val lineLength = distanceBetweenPoints(start, end)

    if (lineLength == 0.0) {
        return distanceBetweenPoints(p, start)
    }

    val u = ((p.latitude - start.latitude) * (end.latitude - start.latitude) +
            (p.longitude - start.longitude) * (end.longitude - start.longitude)) / (lineLength * lineLength)

    if (u < 0 || u > 1) {
        val closestPoint = if (u > 0) end else start
        return distanceBetweenPoints(p, closestPoint)
    }

    val intersection = Point(
        start.latitude + u * (end.latitude - start.latitude),
        start.longitude + u * (end.longitude - start.longitude)
    )

    return distanceBetweenPoints(p, intersection)
}

fun distanceBetweenPoints(p1: Point, p2: Point): Double {
    // 使用简化的球面距离计算
    val radius = 6371.0 // 地球半径，单位：千米
    val dLat = degToRad(p2.latitude - p1.latitude)
    val dLon = degToRad(p2.longitude - p1.longitude)

    val a = Math.sin(dLat / 2).pow(2) +
            Math.cos(degToRad(p1.latitude)) * Math.cos(degToRad(p2.latitude)) *
            Math.sin(dLon / 2).pow(2)

    val c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))

    return radius * c
}

fun degToRad(deg: Double): Double {
    return deg * (Math.PI / 180)
}

fun main() {
    // 示例经纬度轨迹点
    val trajectory = listOf(
        Point(38.115556, 13.361389),
        Point(37.502669, 15.087269),
        Point(38.115556, 13.361389),
        Point(37.502669, 15.087269)
        // 添加更多轨迹点...
    )

    // 误差阈值，根据实际情况调整
    val tolerance = 1.0

    // 应用Douglas-Peucker算法进行轨迹简化
    val simplifiedTrajectory = simplifyDouglasPeucker(trajectory, tolerance)

    // 输出简化后的轨迹点
    println("简化后的轨迹点：$simplifiedTrajectory")
}

```

在这个 Kotlin 示例中，`simplifyDouglasPeucker` 函数是 Douglas-Peucker 算法的实现，用于根据给定的误差阈值 `tolerance` 对轨迹进行简化。同样，你可能需要根据实际需求调整 `tolerance` 和其他参数。