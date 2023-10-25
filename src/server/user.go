package main

import (
	"fmt"
	json "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// 粉丝的权限赋值
const (
	PermissionFlagArticle = 1 << iota // a
	PermissionFlagGps                 // p
	PermissionFlagTrack               // t
)

const PermissionAll = PermissionFlagArticle | PermissionFlagGps | PermissionFlagTrack

// 这里用于存储用户的基本信息，不经常改变的哪些
type UserInfo struct {
	Id      int64  `json:"id"`
	Phone   string `json:"phone"`
	Icon    string `json:"icon"`
	Name    string `json:"name"`
	Nick    string `json:"nick"`
	Email   string `json:"email"`
	Pwd     string `json:"pwd"`
	TempPwd string `json:"temppwd"`
	Region  string `json:"region"`
	Ipv4    string `json:"ipv4"`
	WxId    string `json:"wxid"`
	Age     int8   `json:"age"`
	Gender  int8   `json:"gender"`
	Tm      string `json:"tm"`
	Session int64  `json:"sid"`
}

// 这里用于存储所有用户的动态信息
// up+ id
type UserDetail struct {
	Id     int64  `json:"id"`
	Pt     string `json:"pt"`
	Tm     string `json:"tm"`
	OnLine bool   `json:"online"`
}

// 用于存储用户的会话信息 s+sid
type UserSession struct {
	Sid int64  `json:"sid"`
	Id  int64  `json:"id"`
	Tm  string `json:"tm"`
	Key string `json:"key"`
}

type GroupMember struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Nick  string `json:"nick"`
	GNick string `json:"gnick"`
	Permi int64  `json:"permi"` //授权权限
}

// 存储自己关注的用户，记录是否对对方关注
type FriendInfo struct {
	Id       int64  `json:"id"`
	Icon     string `json:"icon"`
	CreateTm int64  `json:"ctm"`

	Alias   string `json:"alias"`
	Label   string `json:"label"`
	Visible int32  `json:"visible"`
	Block   bool   `json:"block"`
}

// 为了减少传输量
type FunInfo struct {
	Id   int64 `json:"id"`
	Mask int64 `json:"mask"`
}

// 传递的用户权限更改的参数
type FriendPermissinParam struct {
	Id  int64  `json:"id"`
	Fid int64  `json:"fid"`
	Sid int64  `json:"sid"`
	Opt string `json:"opt"`
}

func NewFriendInfo(fid string) *FriendInfo {
	friend := FriendInfo{}
	//friend.IsFriend = false
	friend.Id, _ = strconv.ParseInt(fid, 10, 64)
	friend.CreateTm = time.Now().UnixMilli()
	friend.Visible = PermissionAll
	friend.Block = false
	friend.Label = ""
	return &friend
}

func FriendListToString(lst []*FriendInfo) (string, error) {
	str, err := json.Marshal(lst)
	return string(str), err
}
func UserInfoFormJson(strJson string) (*UserInfo, error) {
	user := NewUserInfo()
	err := json.Unmarshal([]byte(strJson), user)
	return user, err
}

func UserInfoFormBytes(strJson []byte) (*UserInfo, error) {
	user := NewUserInfo()
	err := json.Unmarshal(strJson, user)
	return user, err
}

// 创建一个组，如果是id==0; 否则添加成员到组
type GroupUsers struct {
	GroupId int64    `json:"group_id"`
	Owner   string   `json:"id"`
	List    []string `json:"list"`
}

type ShortMsg struct {
	MsgType string
	GId     string
	FromId  string
	ToId    string
	Tm      int64
	SendSeq int64
	SubType string
	Mesage  string
}

//func GetRandom64() int64 {
//	n := rand.Int63()
//	return n
//}

// 更新关注者的备注时候使用
func (info *FriendInfo) UpdateFrindInfo(newInfo *FriendInfo) {
	info.Alias = newInfo.Alias
	info.Label = newInfo.Label
	info.Visible = newInfo.Visible
	info.Block = newInfo.Block
}

// //////////////////////////////////////////////////////
func (member *GroupMember) GroupMemberToString() (string, error) {
	data, err := json.Marshal(member)
	if err != nil {
		fmt.Printf("序列化错误 err=%v\n", err)
		return "", err
	}
	str := string(data)
	//fmt.Println("序列化后: ", str)
	return str, err
}

func NewUserInfo() *UserInfo {
	user := UserInfo{1000, "0", "sys:icon/1.jpg", "momo", "momo", "momo@local.com",
		"123456", "", "unknow", "0.0.0.0",
		"0", 1, 1, "", 0}
	return &user
}
func (user *UserInfo) UserInfoToString() (string, error) {
	data, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("序列化错误 err=%v\n", err)
		return "", err
	}
	str := string(data)
	//fmt.Println("序列化后: ", str)
	return str, err
}
func UserFromString(strJson string) (*UserInfo, error) {
	user := UserInfo{}
	err := json.Unmarshal([]byte(strJson), &user)
	if err != nil {
		return &user, err
	}
	return &user, nil
}

// 从指针或者结构体转为map类型
func AnyToMap[T any](item T) map[string]interface{} {
	t := reflect.TypeOf(item)
	v := reflect.ValueOf(item)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		fmt.Println("不是struct类型")
	}
	fmt.Printf("name:'%v' kind:'%v'\n", t.Name(), t.Kind())

	var data = make(map[string]interface{})

	//var fTypeK reflect.Kind //其实用fType足够判断类型了，用Kind是个人喜好。
	//var fName string
	//var fType string
	//var fValue reflect.Value

	for i := 0; i < t.NumField(); i++ {

		//fName = t.Field(i).Name
		//fType = t.Field(i).Type.Name()
		//fValue = v.FieldByName(t.Field(i).Name)
		//fTypeK = t.Field(i).Type.Kind()
		//此处可得到所有字段的基本信息
		//fmt.Printf("%s		%s		%s\n", fName, fType, fValue)

		key := strings.ToLower(t.Field(i).Name)
		//value := InterfaceToString(v.Field(i).Interface())
		data[key] = v.Field(i).Interface()
	}
	return data
}

// 从map转为希望的类型，返回指针
// 用法如下： err := AnyFromMapString[UserInfo](data1, user1)
func AnyFromMapString[T any](data map[string]string, item *T) error {
	//t := reflect.TypeOf(item).Elem()
	v := reflect.ValueOf(item).Elem() // 注意，这里必须这样写，使用指针，否则无法赋值

	var fName string
	var keyName string
	var fType reflect.Kind
	var id int
	var err error
	var id64 int64
	var id8 int8

	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i) // a reflect.StructField
		fName = fieldInfo.Name
		fType = fieldInfo.Type.Kind()
		keyName = strings.ToLower(fName)
		if strValue, ok := data[keyName]; ok {
			//fmt.Println("==", fName, fType, strValue, keyName)
			if fType == reflect.String {
				v.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(strValue))
				// 这里不能使用 v.Field(i)
			} else if fType == reflect.Int64 {
				if id, err = strconv.Atoi(strValue); err == nil {
					id64 := int64(id)
					v.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(id64))
					//fmt.Println(fName, fType, strValue, keyName)
				} else {
					fmt.Println(id64, err)
				}
			} else if fType == reflect.Int8 {
				if id, err = strconv.Atoi(strValue); err == nil {
					id8 = int8(id)
					v.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(id8))
					//fmt.Println(fName, fType, strValue, keyName)
				} else {
					fmt.Println(err)
				}
			}
		}

	}
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func AnyFromMap[T any](data map[string]interface{}, item *T) error {
	//t := reflect.TypeOf(item)
	err := mapstructure.Decode(data, item)
	return err
}

// 这里与json的TAG不一样，直接使用名字小写作为键值
func UserFromMap(data map[string]interface{}) (*UserInfo, error) {
	user := UserInfo{}
	err := mapstructure.Decode(data, &user)
	if err != nil {
		fmt.Println(err.Error())
	}

	return &user, nil
}

func UserFromMapString(data map[string]string) (*UserInfo, error) {
	user := NewUserInfo()
	t := reflect.TypeOf(*user)
	v := reflect.ValueOf(user).Elem() // 注意，这里必须这样写，使用指针，否则无法赋值

	var fName string
	var keyName string
	var fType reflect.Kind
	var id int
	var err error
	var id64 int64
	var id8 int8

	for i := 0; i < t.NumField(); i++ {
		fieldInfo := t.Field(i) // a reflect.StructField
		fName = fieldInfo.Name
		fType = fieldInfo.Type.Kind()
		keyName = strings.ToLower(fName)
		if strValue, ok := data[keyName]; ok {
			//fmt.Println("==", fName, fType, strValue, keyName)
			if fType == reflect.String {
				v.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(strValue))
				// 这里不能使用 v.Field(i)
			} else if fType == reflect.Int64 {
				if id, err = strconv.Atoi(strValue); err == nil {
					id64 := int64(id)
					v.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(id64))
					//fmt.Println(fName, fType, strValue, keyName)
				} else {
					fmt.Println(id64, err)
				}
			} else if fType == reflect.Int8 {
				if id, err = strconv.Atoi(strValue); err == nil {
					id8 = int8(id)
					v.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(id8))
					//fmt.Println(fName, fType, strValue, keyName)
				} else {
					fmt.Println(err)
				}
			}
		}

	}
	if err != nil {
		fmt.Println(err.Error())
	}

	return user, nil
}

// 使用反射序列化
func (user *UserInfo) UserInfoToMap() map[string]interface{} {
	t := reflect.TypeOf(*user)
	v := reflect.ValueOf(*user)

	var data = make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		key := strings.ToLower(t.Field(i).Name)
		//value := InterfaceToString(v.Field(i).Interface())
		data[key] = v.Field(i).Interface()
	}
	return data

}

func GroupMemberFromString(strJson string) (*GroupMember, error) {
	mem := GroupMember{}
	err := json.Unmarshal([]byte(strJson), &mem)
	if err != nil {
		return &mem, err
	}
	return &mem, nil
}

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
func RegistUser(rtype string, name string, pwd string, ip string) (*UserInfo, error) {
	id := redisCli.getNextUserId()
	if id <= 1000 {
		return nil, fmt.Errorf("get new user id error")
	}

	user := NewUserInfo()
	user.Id = id
	user.Name = name
	user.Nick = name
	user.Pwd = pwd
	user.TempPwd = ""
	user.Ipv4 = ip
	user.Region = "未设定"

	user.Tm = GetNowTimeString()
	err := redisCli.SetUserInfo(user)
	return user, err
}

// 检查sessionId 是否存在
func CheckUserSession(sid string, id string) bool {
	sess, err := redisCli.FindUserSession(sid)
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
func CreateUserSession(uid string) (*UserSession, error) {
	sess := UserSession{}
	sess.Id, _ = strconv.ParseInt(uid, 10, 64)
	sess.Sid = GetRandom64()
	sess.Tm = GetNowTimeString()
	sess.Key = "nil"
	err := redisCli.SetUserSession(&sess)
	return &sess, err
}

// 使用最基础的用户名口令登录
// 先检查用户名口令，然后检查会话ID
func LoginUserBase(uid string, pwd string) (*UserSession, error) {
	tid, _ := strconv.ParseInt(uid, 10, 64)
	user, err := redisCli.FindUserById(uid)
	if err != nil {
		return nil, err
	}
	if user.Id != tid {
		return nil, fmt.Errorf("can't find user with uid %s", uid)
	}
	if user.Pwd != pwd {
		return nil, fmt.Errorf("user with uid %s, pwd is not correct", uid)
	}

	return CreateUserSession(uid)
}

// 使用最基础的ID和会话ID来确认是否登录，尽量少传递口令，在加密模式下，减少秘钥交换次数
func LoginUserSession(uid string, sid string) (*UserSession, error) {
	ret := CheckUserSession(sid, uid)
	if ret {
		idt, _ := strconv.ParseInt(uid, 10, 64)
		sidt, _ := strconv.ParseInt(sid, 10, 64)
		sess := UserSession{sidt, idt, GetNowTimeString(), ""}
		return &sess, nil
	}

	return nil, fmt.Errorf("can't find sid with uid %s, %s", sid, uid)
}

// 根据某个ID来搜索用户
func SearchUser(uid string, fid string) (*UserInfo, error) {
	user, err := redisCli.FindUserById(fid)
	if user != nil {
		user.Pwd = ""
		user.TempPwd = ""

		// 这里需要判断权限，是否需要屏蔽其他的信息
		user.Email = ""
		user.Phone = ""
	}
	return user, err
}

// 仅仅用来设置基本信息，包包括手机和邮箱, wxid
func SetUserBaseInfo(info *UserInfo) error {
	uid := strconv.FormatInt(info.Id, 10)
	user, err := redisCli.FindUserById(uid)
	if err != nil {
		return err
	}
	if user.Id != info.Id {
		return ErrNoUser
	}

	if len(info.Icon) > 0 {
		user.Icon = info.Icon
	}
	if len(info.Name) > 0 {
		user.Name = info.Name
	}
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

	err = redisCli.SetUserInfo(user)
	return err

}

// 关注对方，
// 将对方添加到自己的列表中,同时将自己添加到对方的粉丝列表
func AddUserFollow(fid string, uid string) error {

	// 检查用户，并添加到自己的关注列表
	user, err := SearchUser(fid, uid)
	if err != nil {
		return fmt.Errorf("can't find friend by id %", fid)
	}

	friend := NewFriendInfo(fid)
	friend.Icon = user.Icon
	friend.Alias = user.Nick
	err = redisCli.SetUserFollow(uid, fid, friend)

	err = redisCli.CreateUserFun(uid, fid)

	return err
}

// 设置一些备注和分组的信息
func SetUserFollowInfo(fid string, uid string, info *FriendInfo) error {
	oldFriendInfo, err := redisCli.GetUserFollow(uid, fid)
	oldFriendInfo.UpdateFrindInfo(info)
	err = redisCli.SetUserFollow(uid, fid, oldFriendInfo)
	return err
}

// 获取关注列表
func GetUserFollowList(uid string, off int, count int) ([]*FriendInfo, error) {
	return redisCli.GetUserFollows(uid)
}

// 获取粉丝列表
func GetUserFunList(uid string, off int, count int) ([]FunInfo, error) {
	return redisCli.GetUserFuns(uid)
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
		mask = mask | PermissionFlagArticle
	}

	if strings.ContainsAny(opt, "pP") {
		mask = mask | PermissionFlagGps
	}

	if strings.ContainsAny(opt, "tT") {
		mask = mask | PermissionFlagTrack
	}

	err := redisCli.UpdateUserFunPermission(uid, funid, mask, set)
	return err
}

// 检查权限
func CheckUserPermission(uid string, fid string) bool {
	if uid == fid {
		return true
	}

	mask := int64(PermissionFlagGps)

	// 这里应该将fid作为用户，检查用户在粉丝中的权限
	return redisCli.CheckUserFunPermission(fid, uid, mask)
}
