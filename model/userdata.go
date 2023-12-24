package model

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
	PermissionFlagShow                // mark show
)

const PermissionAll = PermissionFlagArticle | PermissionFlagGps | PermissionFlagTrack

// 注意：这里必须使用小写的对应的名字，因为反序列化是手写的，使用小写字幕
// 这里用于存储用户的基本信息，不经常改变的哪些
type UserInfo struct {
	Id        int64  `json:"id"`
	Phone     string `json:"phone"`
	Icon      string `json:"icon"`
	Name      string `json:"name"`
	Nick      string `json:"nick"`
	Email     string `json:"email"`
	Pwd       string `json:"pwd"`
	TempPwd   string `json:"temppwd"`
	Region    string `json:"region"`
	Ipv4      string `json:"ipv4"`
	WxId      string `json:"wxid"`
	Age       int8   `json:"age"`
	Gender    int8   `json:"gender"`
	Tm        string `json:"tm"`
	Session   int64  `json:"session"`
	TmLogin   int64  `json:"tmlogin"` // 最后一次登录
	CountFail int8   `json:"countfail"`
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

func NewUserInfo() *UserInfo {
	user := UserInfo{1000, "0", "sys:icon/1.jpg", "momo", "momo", "momo@local.com",
		"123456", "", "unknow", "0.0.0.0",
		"0", 1, 1, "", 0, 0, 0}
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
	var id32 int

	for i := 0; i < t.NumField(); i++ {
		fieldInfo := t.Field(i) // a reflect.StructField
		fName = fieldInfo.Name
		fType = fieldInfo.Type.Kind()
		keyName = strings.ToLower(fName)
		//fmt.Println("=====>", fName, fType)
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
			} else if fType == reflect.Int {
				if id, err = strconv.Atoi(strValue); err == nil {
					id32 = int(id)
					v.FieldByName(fieldInfo.Name).Set(reflect.ValueOf(id32))
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

func (info *FriendInfo) UpdateFrindInfo(newInfo *FriendInfo) {
	info.Alias = newInfo.Alias
	info.Label = newInfo.Label
	info.Visible = newInfo.Visible
	info.Block = newInfo.Block
}
