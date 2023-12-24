package service

import (
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
	"zhituBackend/common"
	"zhituBackend/db"
	"zhituBackend/model"
)

// 用户登录计数器计数器
type UserLoginCounter struct {
	Uid       int64
	TmLast    int64
	FailCount int
}

// 定义一个全局变量
var MapOfAppVistedCount common.ConcurrentMap[int64, *UserLoginCounter]

// 自动初始化
func init() {
	MapOfAppVistedCount = common.NewConcurrentMap[int64, *UserLoginCounter]()
}

func shouldBlock(uid string) bool {
	id, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		return false
	}

	counter, ok := MapOfAppVistedCount.Get(id)
	if ok == false {
		return false
	}

	delta := (time.Now().UnixMilli() - counter.TmLast) / (1000 * 60)

	// 失败3次以上
	if counter.FailCount >= 3 && delta < 5 {
		common.Logger.Info("user login block", zap.String("uid", uid))
		return true
	}

	return false
}

// 会在更新时候加锁期间回调
// 没有值，则设置；如果有，则更新; 需要设置到map的部分通过返回值传递过去
func appAddCallBack(exist bool, valueInMap *UserLoginCounter, newValue *UserLoginCounter) *UserLoginCounter {
	if exist == false {
		return newValue
	} else {
		valueInMap.TmLast = newValue.TmLast
		if newValue.FailCount == 0 {
			valueInMap.FailCount = 0
		} else {
			valueInMap.FailCount += newValue.FailCount
		}

		return valueInMap
	}
}

func setUserLoginResult(uid string, ok bool) bool {

	id, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		return false
	}

	counter := UserLoginCounter{id, time.Now().UnixMilli(), 0}
	if !ok {
		counter.FailCount = 1
	}

	res := MapOfAppVistedCount.Upsert(id, &counter, appAddCallBack)
	common.Logger.Info("user login filter result:",
		zap.String("id", uid),
		zap.Bool("result", ok),
		zap.Int("failcount", res.FailCount))
	return true
}

// 更新关注者的备注时候使用

// //////////////////////////////////////////////////////

// 由基本信息转为快速信息
//func (user *UserInfo) UserInfoToMemberInfo() (*GroupMember, error) {
//	if user == nil {
//		return nil, nil
//	}
//	member := GroupMember{}
//	member.Id = user.Id
//	member.Phone = user.Phone
//	member.Nick = user.Nick
//	member.GNick = user.Nick
//	member.Region = user.Region
//	member.Ipv4 = user.Ipv4
//	member.Gender = user.Gender
//	member.Age = user.Age
//
//	return &member, nil
//}

// 第一阶段是处理类型1，
// 先获取一个
func RegistUser(rtype string, name string, pwd string, ip string) (*model.UserInfo, error) {
	id := db.RedisCli.GetNextUserId()
	if id <= 1000 {
		return nil, fmt.Errorf("get new user id error")
	}

	user := model.NewUserInfo()
	user.Id = id
	user.Name = name
	user.Nick = name
	user.Pwd = pwd
	user.TempPwd = ""
	user.Ipv4 = ip
	user.Icon = fmt.Sprintf("sys:%d", id%35+1)
	user.Region = "未设定"

	user.Tm = common.GetNowTimeString()
	err := db.RedisCli.SetUserInfo(user)
	return user, err
}

// 检查sessionId 是否存在
func CheckUserSession(sid string, id string) bool {
	sess, err := db.RedisCli.FindUserSession(sid)
	if err != nil {
		return false
	}
	if (sess.Sid == 0) || (sess.Id == 0) {
		return false
	}
	temp, _ := strconv.ParseInt(id, 10, 64)
	if temp != sess.Id {
		return false
	}
	return true
}

// 登录后需要设置一个用户的会话键
func CreateUserSession(uid string) (*model.UserSession, error) {
	sess := model.UserSession{}
	sess.Id, _ = strconv.ParseInt(uid, 10, 64)
	sess.Sid = common.GetRandom64()
	sess.Tm = common.GetNowTimeString()
	sess.Key = "nil"
	err := db.RedisCli.SetUserSession(&sess)
	return &sess, err
}

// 使用最基础的用户名口令登录
// 先检查用户名口令，然后检查会话ID
func LoginUserBase(uid string, pwd string) (*model.UserSession, error) {

	// 过滤3次错误的，需要等待5分钟
	bBlock := shouldBlock(uid)
	if bBlock {
		return nil, fmt.Errorf("user with uid %s, pwd is not correct for 3 times, wait 5 minute", uid)
	}

	tid, _ := strconv.ParseInt(uid, 10, 64)
	user, err := db.RedisCli.FindUserById(uid)
	//jsonUser, _ := user.UserInfoToString()
	//common.Logger.Debug("login base, find user", zap.String("user", jsonUser))
	if err != nil {
		return nil, err
	}
	if user.Id != tid {
		return nil, fmt.Errorf("can't find user with uid %s", uid)
	}
	if user.Pwd != pwd {
		setUserLoginResult(uid, false)
		db.RedisCli.SetUserloginResult(user, false)
		return nil, fmt.Errorf("user with uid %s, pwd is not correct", uid)
	}

	setUserLoginResult(uid, true)
	db.RedisCli.SetUserloginResult(user, true)
	common.Logger.Info("user login ok", zap.String("name", uid))
	return CreateUserSession(uid)
}

// 使用最基础的ID和会话ID来确认是否登录，尽量少传递口令，在加密模式下，减少秘钥交换次数
func LoginUserSession(uid string, sid string) (*model.UserSession, error) {
	ret := CheckUserSession(sid, uid)
	if ret {
		idt, _ := strconv.ParseInt(uid, 10, 64)
		sidt, _ := strconv.ParseInt(sid, 10, 64)
		sess := model.UserSession{sidt, idt, common.GetNowTimeString(), ""}
		return &sess, nil
	}

	return nil, fmt.Errorf("can't find sid with uid %s, %s", sid, uid)
}

// 根据某个ID来搜索用户
func SearchUser(uid string, fid string) (*model.UserInfo, error) {
	user, err := db.RedisCli.FindUserById(fid)
	if user != nil {
		user.Pwd = ""
		user.TempPwd = ""

		tempId := strconv.FormatInt(user.Id, 10)

		if tempId != uid {
			// 这里需要判断权限，是否需要屏蔽其他的信息
			user.Email = ""
			user.Phone = ""
		}
	}

	return user, err
}

// 仅仅用来设置基本信息，包包括手机和邮箱, wxid
func SetUserBaseInfo(info *model.UserInfo) error {
	uid := strconv.FormatInt(info.Id, 10)
	user, err := db.RedisCli.FindUserById(uid)
	if err != nil {
		return err
	}
	if user.Id != info.Id {
		return model.ErrNoUser
	}

	if len(info.Icon) > 0 {
		user.Icon = info.Icon
	}
	//if len(info.Name) > 0 {
	//	user.Name = info.Name
	//}
	if len(info.Nick) > 0 {
		user.Nick = info.Nick
	}

	if info.Age > 0 {
		user.Age = info.Age
	}
	if info.Gender > -1 {
		user.Gender = info.Gender
	}
	if len(info.Pwd) > 0 {
		user.Pwd = info.Pwd
	}

	if len(info.Phone) > 0 {
		user.Phone = info.Phone
	}

	if len(info.Email) > 0 {
		user.Email = info.Email
	}

	fmt.Println(info)

	err = db.RedisCli.SetUserInfo(user)
	return err

}

// 关注对方，
// 将对方添加到自己的列表中,同时将自己添加到对方的粉丝列表
func AddUserFollow(fid string, uid string) error {

	// 检查用户，并添加到自己的关注列表
	user, err := SearchUser(uid, fid)
	if err != nil {
		return fmt.Errorf("can't find friend by id %", fid)
	}

	friend := model.NewFriendInfo(fid)
	friend.Icon = user.Icon
	friend.Alias = fmt.Sprintf("%s (%s)", user.Name, user.Nick)
	err = db.RedisCli.SetUserFollow(uid, fid, friend)

	err = db.RedisCli.CreateUserFun(uid, fid)

	return err
}

// 设置一些备注和分组的信息
func SetUserFollowInfo(fid string, uid string, info *model.FriendInfo) error {
	oldFriendInfo, err := db.RedisCli.GetUserFollow(uid, fid)
	oldFriendInfo.UpdateFrindInfo(info)
	err = db.RedisCli.SetUserFollow(uid, fid, oldFriendInfo)
	return err
}

// 删除关注，或者设置是否显示
func SetUserFollowSimple(fid string, uid string, param string) error {
	err := db.RedisCli.SetFollowUserInfo(uid, fid, param)
	return err
}

// 获取关注列表
func GetUserFollowList(uid string, off int, count int) ([]*model.FriendInfo, error) {
	return db.RedisCli.GetUserFollows(uid)
}

// 获取粉丝列表
func GetUserFunList(uid string, off int, count int) ([]model.FunInfo, error) {
	return db.RedisCli.GetUserFuns(uid)
}

// 设置粉丝权限
// +apt
// -apt
func SetUserFunPermission(uid string, sid string, funid string, opt string) error {

	set := true
	mask := int64(0)
	// 设置了减号默认认为是清楚，否则就是开放权限
	if strings.HasPrefix(opt, "-") {
		set = false
	}
	if strings.ContainsAny(opt, "aA") {
		mask = mask | model.PermissionFlagArticle
	}

	if strings.ContainsAny(opt, "pP") {
		mask = mask | model.PermissionFlagGps
	}

	if strings.ContainsAny(opt, "tT") {
		mask = mask | model.PermissionFlagTrack
	}

	err := db.RedisCli.UpdateUserFunPermission(uid, funid, mask, set)
	return err
}

// 检查权限
func CheckUserPermission(uid string, fid string) bool {
	if uid == fid {
		return true
	}

	// 如果对方关注你了，则你可以查看对方位置
	ret := db.RedisCli.AIsFollowingB(fid, uid)
	if ret {
		return true
	}

	mask := int64(model.PermissionFlagGps)
	// 这里应该将fid作为用户，检查用户在粉丝中的权限
	return db.RedisCli.CheckUserFunPermission(fid, uid, mask)
}
