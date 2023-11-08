package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	json "github.com/json-iterator/go"
	"math"
	"net"
	"sort"
	"strconv"
	"time"
)

var ctx = context.Background()

type RedisClient struct {
	Db *redis.Client
}

// host should be like this: "10.128.5.73:6379"
func NewRedisClient(host string, pwd string) *RedisClient {
	cli := RedisClient{}
	err, redisdb := initRedis(host, pwd)
	if err != nil {
		fmt.Printf("connect redis failed! err : %v\n", err)
		return nil
	}

	cli.Db = redisdb
	return &cli
}

func (cli *RedisClient) Close() {
	cli.Close()
}

// https://blog.csdn.net/weixin_45901764/article/details/117226225
func initRedis(addr string, password string) (err error, redisdb *redis.Client) {
	redis_opt := redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
		Network:  "tcp", //网络类型，tcp or unix，默认tcp

		//连接池容量及闲置连接数量
		PoolSize:     150, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
		MinIdleConns: 10,  //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；。

		//超时
		DialTimeout:  5 * time.Second, //连接建立超时时间，默认5秒。
		ReadTimeout:  3 * time.Second, //读超时，默认3秒， -1表示取消读超时
		WriteTimeout: 3 * time.Second, //写超时，默认等于读超时
		PoolTimeout:  4 * time.Second, //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒。

		//闲置连接检查包括IdleTimeout，MaxConnAge
		IdleCheckFrequency: 60 * time.Second, //闲置连接检查的周期，默认为1分钟，-1表示不做周期性检查，只在客户端获取连接时对闲置连接进行处理。
		IdleTimeout:        5 * time.Minute,  //闲置超时，默认5分钟，-1表示取消闲置超时检查
		MaxConnAge:         0 * time.Second,  //连接存活时长，从创建开始计时，超过指定时长则关闭连接，默认为0，即不关闭存活时长较长的连接

		//命令执行失败时的重试策略
		MaxRetries:      0,                      // 命令执行失败时，最多重试多少次，默认为0即不重试
		MinRetryBackoff: 8 * time.Millisecond,   //每次计算重试间隔时间的下限，默认8毫秒，-1表示取消间隔
		MaxRetryBackoff: 512 * time.Millisecond, //每次计算重试间隔时间的上限，默认512毫秒，-1表示取消间隔

		//可自定义连接函数
		Dialer: func() (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Minute,
			}
			return netDialer.Dial("tcp", addr)
		},

		//钩子函数
		OnConnect: func(conn *redis.Conn) error { //仅当客户端执行命令时需要从连接池获取连接时，如果连接池需要新建连接时则会调用此钩子函数
			//fmt.Printf("conn=%v\n", conn)
			return nil
		},
	}
	redisdb = redis.NewClient(&redis_opt)
	// 判断是否能够链接到数据库
	pong, err := redisdb.Ping().Result()
	if err != nil {
		fmt.Println(pong, err)
	}

	//printRedisPool(redisdb.PoolStats())
	return err, redisdb
}

func printRedisPool(stats *redis.PoolStats) {
	fmt.Printf("Hits=%d Misses=%d Timeouts=%d TotalConns=%d IdleConns=%d StaleConns=%d\n",
		stats.Hits, stats.Misses, stats.Timeouts, stats.TotalConns, stats.IdleConns, stats.StaleConns)
}

func printRedisOption(opt *redis.Options) {
	fmt.Printf("Network=%v\n", opt.Network)
	fmt.Printf("Addr=%v\n", opt.Addr)
	fmt.Printf("Password=%v\n", opt.Password)
	fmt.Printf("DB=%v\n", opt.DB)
	fmt.Printf("MaxRetries=%v\n", opt.MaxRetries)
	fmt.Printf("MinRetryBackoff=%v\n", opt.MinRetryBackoff)
	fmt.Printf("MaxRetryBackoff=%v\n", opt.MaxRetryBackoff)
	fmt.Printf("DialTimeout=%v\n", opt.DialTimeout)
	fmt.Printf("ReadTimeout=%v\n", opt.ReadTimeout)
	fmt.Printf("WriteTimeout=%v\n", opt.WriteTimeout)
	fmt.Printf("PoolSize=%v\n", opt.PoolSize)
	fmt.Printf("MinIdleConns=%v\n", opt.MinIdleConns)
	fmt.Printf("MaxConnAge=%v\n", opt.MaxConnAge)
	fmt.Printf("PoolTimeout=%v\n", opt.PoolTimeout)
	fmt.Printf("IdleTimeout=%v\n", opt.IdleTimeout)
	fmt.Printf("IdleCheckFrequency=%v\n", opt.IdleCheckFrequency)
	fmt.Printf("TLSConfig=%v\n", opt.TLSConfig)

}

// 使用普通的key保存一个值，每次都增加一个值，用于计算新增用户的ID
func (cli *RedisClient) getNextUserId() (id int64) {
	if cli == nil {
		return -1
	}
	idCmd := cli.Db.Incr("users_id")
	return idCmd.Val()
}

func (cli *RedisClient) getNextGroupId() (id int64) {
	if cli == nil {
		return -1
	}
	idCmd := cli.Db.Incr("groups_id")
	return idCmd.Val()
}

func (cli *RedisClient) FindUserById(fid string) (*UserInfo, error) {
	tblName := "u" + fid
	//fmt.Println(tblName)
	data, err := cli.Db.HGetAll(tblName).Result()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	user, err := UserFromMapString(data)
	//fmt.Println(data)
	//fmt.Println(user)
	return user, err

}

// 保存一个完整的user
func (cli *RedisClient) SetUserInfo(user *UserInfo) error {
	tblName := "u" + strconv.FormatInt(user.Id, 10)
	table := user.UserInfoToMap()
	ret, err := cli.Db.HMSet(tblName, table).Result()
	fmt.Println(ret)
	return err
}

// 创建群或者添加用户信息
//func (cli *RedisClient) AddUsersToGroup(members *GroupUsers) (id int64) {
//	if members.GroupId < 1 {
//		members.GroupId = cli.getNextGroupId()
//	}
//	tblName := "gh_" + strconv.FormatInt(members.GroupId, 10)
//	table := make(map[string]interface{})
//	memList := make([]interface{}, 0, 5)
//
//	// 提取所有的基本信息，然后转换为群属性，保存到个人部分
//	for _, phone := range members.List {
//		memList = append(memList, phone)
//		user, err := cli.FindUserById(phone)
//		if user == nil || err != nil {
//			continue
//		}
//		member, err := user.UserInfoToMemberInfo()
//		if err != nil {
//			continue
//		}
//		strJson, err := member.GroupMemberToString()
//		if err != nil {
//			continue
//		}
//		table[phone] = strJson
//		// 用户的组信息中追加标记组, 这里使用了SET
//		keyName := "uing_" + phone
//		cli.Db.SAdd(keyName, members.GroupId)
//	}
//
//	// 哈希是为了保存群成员基础信息
//	cli.Db.HMSet(tblName, table)
//
//	// set是为了快速查找所有用户的账号，
//	tblName = "gs_" + strconv.FormatInt(members.GroupId, 10)
//	cli.Db.SAdd(tblName, memList...)
//	return members.GroupId
//}

// 存储用户的会话信息,比如 s1001
func (cli *RedisClient) SetUserSession(sess *UserSession) error {
	tblName := "s" + strconv.FormatInt(sess.Sid, 10)
	table := AnyToMap[*UserSession](sess)
	ret, err := cli.Db.HMSet(tblName, table).Result()
	fmt.Println(ret)

	// 设置该键超时的时长
	cli.Db.Expire(tblName, 7*24*time.Hour)
	return err
}

// 检索用户的会话ID
func (cli *RedisClient) FindUserSession(sid string) (*UserSession, error) {
	tblName := "s" + sid
	data, err := cli.Db.HGetAll(tblName).Result()
	if err != nil {
		return nil, err
	}
	sess := UserSession{}
	sess.Sid = 0
	sess.Id = 0
	err = AnyFromMapString[UserSession](data, &sess)
	return &sess, err

}

// 关注对方，同时将对方添加到自己关注列表，
// todo: 将自己添加到对方的粉丝列表中
func (cli *RedisClient) SetUserFollow(uid string, fid string, friend *FriendInfo) error {
	data, err := json.Marshal(&friend)
	if err != nil {
		fmt.Printf("序列化错误 err=%v\n", err)
		return fmt.Errorf("can't serilize friend by id %", fid)
	}
	str := string(data)
	key := "uf" + uid
	fmt.Println("设置用户的关注：", key, fid)
	_, err = cli.Db.HSet(key, fid, str).Result()
	return err
}

// 取消关注或者更显关注的信息
// show, ignore, remove
func (cli *RedisClient) setFollowUserInfo(uid string, fid string, param string) error {

	friend := NewFriendInfo(fid)
	tblName := "uf" + uid
	strJson, err := cli.Db.HGet(tblName, fid).Result()

	err = json.Unmarshal([]byte(strJson), friend)
	if err != nil {
		return err
	}

	if param == "ignore" {
		friend.Visible = friend.Visible & (^PermissionFlagShow)
		err = cli.SetUserFollow(uid, fid, friend)
		return err

	} else if param == "show" {
		friend.Visible |= PermissionFlagShow
		err = cli.SetUserFollow(uid, fid, friend)
		return err

	} else if param == "remove" {
		key := "uf" + uid
		_, err = cli.Db.HDel(key, fid).Result()
		return err
	}

	return nil
}

// 检查A的关注列表中是否有B，如果有，则B可以查看A
func (cli *RedisClient) AIsFollowingB(A string, B string) bool {

	tblName := "uf" + A
	_, err := cli.Db.HGet(tblName, B).Result()
	if err == nil {
		return true
	}

	return false
}

// 查询关注好友的相关信息
func (cli *RedisClient) GetUserFollow(uid string, fid string) (*FriendInfo, error) {
	friend := NewFriendInfo(fid)
	tblName := "uf" + uid
	strJson, err := cli.Db.HGet(tblName, fid).Result()

	err = json.Unmarshal([]byte(strJson), friend)
	if err != nil {
		return nil, err
	}

	return friend, err
}

// 这里叫做订阅列表，订阅列表比较简单，就是一个权限掩码
func (cli *RedisClient) CreateUserFun(uid string, fid string) error {

	// 发布者的粉丝列表
	key := "us" + fid
	// 开始时候先设置全都允许的掩码，后面再改
	str := strconv.FormatInt(0xffff, 16)
	_, err := cli.Db.HSet(key, uid, str).Result()
	return err
}

// 获取所有的关注列表
func (cli *RedisClient) GetUserFollows(uid string) ([]*FriendInfo, error) {
	key := "uf" + uid
	data, err := cli.Db.HGetAll(key).Result()
	if err != nil {
		return make([]*FriendInfo, 1), err
	}

	var friendinfo *FriendInfo
	count := len(data)
	lst := make([]*FriendInfo, count)
	index := 0
	for fid, jsonStr := range data {
		friendinfo = NewFriendInfo(fid)
		err = json.Unmarshal([]byte(jsonStr), friendinfo)
		lst[index] = friendinfo
		index++
	}

	return lst, err
}

// 获取当前粉丝列表
func (cli *RedisClient) GetUserFuns(uid string) ([]FunInfo, error) {
	// 发布者的粉丝列表
	key := "us" + uid
	data, err := cli.Db.HGetAll(key).Result()
	if err != nil {
		return make([]FunInfo, 1), err
	}

	var id int64
	count := len(data)
	lst := make([]FunInfo, count)
	index := 0
	for fid, strmask := range data {
		id, _ = strconv.ParseInt(fid, 10, 64)
		//fmt.Println(strmask)
		mask, _ := strconv.ParseInt(strmask, 16, 64)
		//fmt.Println(mask)
		lst[index].Id = id
		lst[index].Mask = mask
		index++
	}

	return lst, err
}

// 设置或者清除某一个标志位
func (cli *RedisClient) UpdateUserFunPermission(uid string, funid string, mask int64, set bool) error {
	key := "us" + uid
	str, err := cli.Db.HGet(key, funid).Result()
	if err != nil {
		return err
	}
	oldmask, err := strconv.ParseInt(str, 16, 64)
	if err != nil {
		return err
	}
	if set {
		oldmask = oldmask | mask
	} else {
		oldmask = oldmask & (^mask)
	}

	str = strconv.FormatInt(oldmask, 16)
	_, err = cli.Db.HSet(key, funid, str).Result()
	return err
}

// 检查uid，fid粉丝列表中的权限
func (cli *RedisClient) CheckUserFunPermission(uid string, fid string, mask int64) bool {
	key := "us" + uid
	str, err := cli.Db.HGet(key, fid).Result()
	if err != nil {
		return false
	}
	oldmask, _ := strconv.ParseInt(str, 16, 64)
	if (mask & oldmask) > 0 {
		return true
	}
	return false
}

// /////////////////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////////////////////////////////
// /////////////////////////////////////////////////////////////////////////////
// up${id}表存储用户的动态信息
// 到user表中查找数据
// up1001
func (cli *RedisClient) FindLastGpx(uid string) (data *GpxData, err error) {
	tblName := "up" + uid
	value, err := cli.Db.HGet(tblName, "pt").Result()

	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, ErrNoValue
	}

	data, err = GpxDataFromJson(value)
	return data, err
}

// 这种方式先从群组读出成员列表，在多读方式直接获取一组数据，这样写的压力小，因为读比写要少
// 备注：但是在集群模式下，由于slot映射，如果不在一个slot上，会执行多次请求，效率会下降，这时需要使用另一种策略，写扩展
// 写扩散也就是用户上报信息时候，同时写到了各个群组中
func (cli *RedisClient) FindGroupLastGpx(gId string) (*GpxDataArray, error) {
	tblName := "ggpx_" + gId
	memberList, err := cli.Db.SMembers(tblName).Result()
	if err != nil {
		return nil, err
	}
	retData := GpxDataArray{}
	retData.Uid = gId
	groupKeys := make([]string, 0, 5)
	for _, phone := range memberList {
		key := "ulast_" + phone
		groupKeys = append(groupKeys, key)
	}
	// 组成员为空，不可能，肯定出错了
	if len(groupKeys) < 1 {
		return nil, err
	}
	dataList, err := cli.Db.MGet(groupKeys...).Result()
	for _, gpxStr := range dataList {
		var gpx *GpxData
		strTemp := InterfaceToString(gpxStr)
		gpx, err = GpxDataFromJson(strTemp)
		if err != nil {
			continue
		}

		retData.DataList = append(retData.DataList, *gpx)
	}

	return &retData, err
}

// 找到分片轨迹表中最后一条数据, 此函数主要是内部使用
// ugpx13800138000_20060102
func (cli *RedisClient) FindPreData(tblNameGpx string) *GpxData {
	// 检查最后一条
	strList, err := cli.Db.LRange(tblNameGpx, 0, 0).Result()
	if err != nil {
		return nil
	}

	if len(strList) == 0 {
		return nil
	}
	//fmt.Println(strList)
	gpx, err := GpxDataFromJson(strList[0])
	if err != nil {
		return nil
	}

	return gpx
}

// 简单的判断是否静止的算法
// 在纬度相等的情况下: 经度每隔0.00001度,距离相差约1米
// 在经度相等的情况下: 纬度每隔0.00001度,距离相差约1.1米
func (cli *RedisClient) isStill(data *GpxData, preData *GpxData) bool {
	if data == nil || preData == nil {
		fmt.Printf("空，空不等")
		return false
	}
	if math.Abs(data.Lon-preData.Lon) > 0.00001*3 {
		fmt.Printf("经度变化")
		return false
	}
	if math.Abs(data.Lat-preData.Lat) > 0.00001*3 {
		fmt.Printf("纬度变化")
		return false
	}
	if data.Speed != 0 {
		fmt.Printf("速度变化")
		return false
	}
	return true
}

// 返回当前日期的条目个数，并且返回当前最后插入的时间戳，便于上传
// 更新2个表，首先到user表比较当查找最近的记录，检查数据是否需要插入，如果需要插入则更新user表，
// 到user表中数据u_13800138000，
// 这里还做一个优化，如果移动距离太小，则认为没有移动，则不写入队列
func (cli *RedisClient) AddGpx(data *GpxData) (count int64, tm int64, err error) {
	// 空指针，无法计算用户
	if data == nil {
		return 0, 0, ErrNilPointer
	}

	// 如果是时间戳是保存过的，则跳过
	gpx, err := cli.FindLastGpx(data.Uid)
	if gpx != nil && data.Tm <= gpx.Tm {
		err = ErrOldValue
		return 0, gpx.Tm, ErrOldValue
	}

	// 先更新最后地点，再添加到队列，如果不能按照事务完成，则轨迹中会缺少数据
	tblNameU := "up" + data.Uid
	strJson, err := data.ToJsonString()
	cli.Db.HSet(tblNameU, "pt", strJson)

	// 这里是为了获取群组信息时候一次性获取
	//keyName := "ulast_" + data.Uid
	//cli.Db.Set(keyName, strJson, 0)

	// 计算当前轨迹表分片名字
	ptTm := time.Unix(data.Tm, 0)
	ptTmStr := ptTm.Format("20060102")
	tblNameGpx := "ugpx" + data.Uid + "_" + ptTmStr

	// 判断是否一直静止，不需要插入
	preData := cli.FindPreData(tblNameGpx)
	//fmt.Println(preData)
	if preData != nil && cli.isStill(data, preData) {
		// 太近，认为没有动，不存
		return 1, data.Tm, nil
	}
	// 左侧添加到列表
	cmd := cli.Db.LPush(tblNameGpx, strJson)
	if preData == nil { // 先建立的表，需要设置超时时间7天
		cli.Db.ExpireAt(tblNameGpx, time.Now().Add(7*24*time.Hour))
	}

	// 添加到列表
	return cmd.Val(), data.Tm, nil
}

// 返回当前日期的条目个数，并且返回当前最后插入的时间戳，便于上传
func (cli *RedisClient) AddGpxDataArray(array *GpxDataArray) (count int64, tm int64, err error) {

	// 空指针，无法计算用户
	if array == nil {
		return 0, 0, ErrNilPointer
	}

	// 先查一下最近的更新数据
	gpx, err := cli.FindLastGpx(array.Uid)

	// 应该先试着排序一下，防止上报数据乱序

	count = 0
	tm = 0
	sort.Sort(array.DataList) // 从小到大排序时间戳，理论上不需要排序

	// 先检查一下是否都太陈旧了
	length := array.DataList.Len()
	lstData := array.DataList[length-1]
	if gpx != nil && lstData.Tm <= gpx.Tm {
		// 已经保存过的了
		return count, gpx.Tm, ErrOldValue

	} else if lstData.Tm < time.Now().Unix()-7*24*3600 {
		// 7天前的数据
		return count, gpx.Tm, ErrOldValue
	} else {
		// 将更新最后时间点
		tblNameU := "up" + array.Uid
		strJson, err := lstData.ToJsonString()
		if err == nil {
			cli.Db.HSet(tblNameU, "pt", strJson)
		} else {
			// 不太可能出现这个

		}

	}

	var nameMap map[string]int64 = make(map[string]int64)
	// todo: 这里应该改为批量插入，速度能快一些
	for index, data := range array.DataList {
		// 值比较陈旧，已经保存过了
		if gpx != nil && data.Tm <= gpx.Tm {
			continue
		}
		// 检测是否过近
		if index > 0 {
			preData := array.DataList[index-1]
			if cli.isStill(&data, &preData) {
				continue
			}
		}

		ptTm := time.Unix(data.Tm, 0)
		ptTmStr := ptTm.Format("20060102")
		tblNameGpx := "ugpx" + array.Uid + "_" + ptTmStr
		if _, ok := nameMap[tblNameGpx]; ok {
			// 存在
			nameMap[tblNameGpx] += 1
		} else {
			nameMap[tblNameGpx] = 1
		}

		strJson, err := data.ToJsonString()
		if err == nil {
			cmd := cli.Db.LPush(tblNameGpx, strJson)
			count = cmd.Val()
			tm = data.Tm

		}

	}

	// 设置各个按日期的分表超时时长，这里应该根据配置来决定
	for k, _ := range nameMap {
		//cli.Db.Expire(k, 60*time.Second)
		cli.Db.Expire(k, 7*24*time.Hour)
	}

	// 列表中的值都太陈旧了
	if gpx != nil && count == 0 {
		return count, gpx.Tm, ErrOldValue
	}

	return count, tm, nil
}

// 按照用户和日期来查找数据，1个list最多也就是3600*24条
func (cli *RedisClient) FindGpxTrack(param *QueryParamTrack) (array *GpxDataArray, err error) {
	tblNameGpx := "ugpx" + param.Fid + "_" + param.DateStr
	cmd := cli.Db.LRange(tblNameGpx, 0, -1)
	strJsonArray, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	// 如果有时间范围，则过滤一下
	//fmt.Println(param.TmStart)
	// 过滤数据
	local, _ := time.LoadLocation("Asia/Shanghai")
	tmDate, _ := time.ParseInLocation("20060102", param.DateStr, local)
	//fmt.Println(tmDate)

	// 如果范围时间戳合法，则过滤，先修正一下结束时间，这个可以不设置
	if param.TmEnd < param.TmStart {
		param.TmEnd = tmDate.Unix() + 24*3600 + 60
	}

	bFilter := false
	// 如果开始时间戳没有正确设置，则不过滤: 应该设置日期，并选择过滤时间，日期是必选的
	if param.TmStart >= tmDate.Unix() && param.TmStart <= (tmDate.Unix()+24*3600) {
		bFilter = true
	}

	//fmt.Println(strJsonArray)
	// 这里的查询结果是时间从近到远
	array = NewGpxDataList()
	array.Uid = param.Fid
	var lastPt *GpxData // 找到时间戳前面的一个点
	lastPt = nil
	for _, str := range strJsonArray {

		gpx, err := GpxDataFromJson(str)
		if err == nil {
			// 如果过滤， 则查看时间戳，否则，直接添加
			if bFilter {
				if param.TmStart <= gpx.Tm && gpx.Tm <= param.TmEnd {
					// 在插入的第一个点前，尝试插入一个之前的点，防止静止不动的情况
					if len(array.DataList) < 1 && lastPt != nil {
						array.DataList = append(array.DataList, *lastPt)
					}
					array.DataList = append(array.DataList, *gpx)
				} else if param.TmStart > gpx.Tm {
					// 当前列表为空，可能一直没有移动，此时需要补充一条
					if len(array.DataList) < 1 {
						lastPt = gpx
					}
					break

				}
			} else { // 否则直接添加
				array.DataList = append(array.DataList, *gpx)
			}

		}
	}

	//fmt.Println(array.ToJsonString())
	return array, nil

}

// /////////////////////////////////////////////////////////////
// for test it only
func (cli *RedisClient) TestAddData() {
	data := GpxData{"13810501031", 40.1, 116.12, 12, 0, 0}
	data.Tm = time.Now().Unix()
	n, lstTm, err := cli.AddGpx(&data)

	array := NewGpxDataList()
	array.Uid = "1001"
	i := 0
	for i < 4 {
		// 保证采集的时间戳都不一样
		time.Sleep(time.Duration(1) * time.Second)
		i++
		gpx := GpxData{}
		gpx.Speed = float64(i)
		gpx.Lat = 40
		gpx.Lon = 116
		gpx.Ele = 10
		gpx.Tm = time.Now().Unix()
		array.DataList = append(array.DataList, gpx)

	}
	fmt.Println("队列长度：", array.DataList.Len())
	n, lstTm, err = cli.AddGpxDataArray(array)

	ptTm := time.Unix(lstTm, 0)
	// 注意，这里需要使用固定的数字格式化
	ptTmStr := ptTm.Format("2006年1月2日 03点04分05秒")

	fmt.Printf("插入数据: 长度%d, 最后时间%s \n", n, ptTmStr)

	lastGpx, err1 := cli.FindLastGpx("13810501031")
	if err1 == nil {
		str, _ := lastGpx.ToJsonString()
		fmt.Println("最后活动点", str)
	} else if err1 == redis.Nil {
		fmt.Println(err.Error())
	}
}
