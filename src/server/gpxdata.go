package main

import (
	"fmt"
	//"encoding/json"
	json "github.com/json-iterator/go"
)

// 这个文件主要是用于定义各种传输的数据结构，使用json编码
// 数据点

type GpxData struct {
	Uid   string  `json:"uid" `
	Lat   float64 `json:"lat" `
	Lon   float64 `json:"lon" `
	Ele   float64 `json:"ele" `
	Speed float64 `json:"speed" `
	Tm    int64   `json:"tm" ` // second time stamp
	TmStr string  `json:"tmStr"`

	Accuracy  float64 `json:"accuracy"`
	Src       string  `json:"src"`
	Direction float64 `json:"direction"`
	City      string  `json:"city"`
	Addr      string  `json:"addr"`
	Street    string  `json:"street"`
	StreetNo  string  `json:"streetNo"`
	//	Rest  int8    `json: "rest" `  // 1：处于静止中， 0，未设置
}

// 重新定义一个类型，用于排序
type GpxDataList []GpxData

// 下面的三个函数必须实现（获取长度函数，交换函数，比较函数（这里比较的是年龄））
func (list GpxDataList) Len() int {
	return len(list)
}
func (list GpxDataList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
func (list GpxDataList) Less(i, j int) bool {
	return list[i].Tm < list[j].Tm
}

// 点的集合
type GpxDataArray struct {
	Uid      string      `json:"uid" `
	DataList GpxDataList `json:"list" `
	//	PrePt    GpxData     `json:"prePt"`
	//列表如果为空时，设置查到的前1个点
}

// 请求轨迹数据的参数，默认只需要提供日期，也可以按照时间戳范围过滤
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
	Fid string `json:"fid" `
}

////////////////////////////////////////////////////////////

// 创建一个空的集合
func NewGpxDataList() *GpxDataArray {
	array := GpxDataArray{}
	array.DataList = make([]GpxData, 0)
	return &array
}

func (list *GpxDataArray) ToJsonString() (str string, err error) {
	//data, err := json.MarshalIndent(&list, "", "    ")
	data, err := json.Marshal(&list)
	if err != nil {
		fmt.Printf("序列化错误 err=%v\n", err)
		return
	}
	str = string(data)
	//fmt.Println("序列化后: ", str)
	return str, err
}
func (list *GpxDataArray) FromJsonString(strJson string) (array *GpxDataArray, err error) {
	if list == nil {
		list = NewGpxDataList()
	}
	err = json.Unmarshal([]byte(strJson), &list)
	if err != nil {
		return nil, err
	}

	return list, err
}

func (gpx *GpxData) ToJsonString() (str string, err error) {
	data, err := json.Marshal(&gpx)
	if err != nil {
		fmt.Printf("序列化错误 err=%v\n", err)
		return
	}
	str = string(data)
	//fmt.Println("序列化后: ", str)
	return str, err
}

func GpxDataFromJson(strJson string) (gpxRet *GpxData, err error) {
	var data GpxData
	err = json.Unmarshal([]byte(strJson), &data)
	if err != nil {
		return nil, err
	}

	return &data, err
}

func (param *QueryParamPoint) QueryParamPointFromJson(strJson string) (err error) {
	err = json.Unmarshal([]byte(strJson), param)

	return err
}

func (param *QueryParamTrack) QueryParamTrackFromJson(strJson string) (err error) {
	err = json.Unmarshal([]byte(strJson), param)

	return err
}

//func (param *QueryParam) ToString(strJson string) (data string, err error) {
//	err = json.Unmarshal([]byte(strJson), param)
//
//	return err
//}
