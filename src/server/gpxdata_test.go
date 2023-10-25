package main

import (
	"fmt"
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	gpx := GpxData{"13800138000", 40.1, 116.12, 12, 0, 0}
	str, err := gpx.ToJsonString()
	//destStr := `{"phone":"13800138000","lat":40.1,"lon":116.12,"ele":12,"tm":0,"speed":0}`

	if err != nil {
		t.Errorf("%s is error", err)
	}
	fmt.Println(str)
}

func TestMarshal1(t *testing.T) {
	gpx := GpxData{}
	gpx.Phone = "13810501031"
	gpx.Tm = time.Now().UnixNano() / 1e9 //  seconds from 1970

	str, err := gpx.ToJsonString()
	//destStr := `{"phone":"13800138000","lat":40.1,"lon":116.12,"ele":12,"tm":0,"speed":0}`

	if err != nil {
		t.Errorf("%s is error", err)
	}
	fmt.Println(str)
}

func TestUnmarshal(t *testing.T) {
	strJson := `{"phone":"13800138000","lat":40.1,"lon":116.12,"ele":12,"tm":0,"speed":0}`
	var gpxData *GpxData
	var err error
	gpxData, err = gpxData.FromJsonString(strJson)
	if err != nil {
		t.Errorf("%s is error", err)
	}
}

func TestUnmarshal1(t *testing.T) {
	strJson := `{"phone":13810501031,"lat":40.1,"lon":116.12,"ele":12,"tm":0,"speed":0}`
	var gpxData *GpxData
	var err error
	gpxData, err = gpxData.FromJsonString(strJson)
	if err != nil {
		t.Errorf("%s is error", err)
	}
}

func TestMarshal2(t *testing.T) {
	list := NewGpxDataList()
	list.Phone = "13800138000"
	i := 1
	for i < 4 {
		gpx := GpxData{}
		gpx.Speed = float64(i)
		gpx.Lat = 20
		i = i + 1

		list.DataList = append(list.DataList, gpx)
	}
	fmt.Println(list.ToJsonString())
}

func TestUnmarshal2(t *testing.T) {
	str := `{"phone":"13800138000","list":[{"lat":20,"lon":0,"ele":0,"speed":2,"tm":0},{"lat":20,"lon":0,"ele":0,"speed":3,"tm":0}]} `
	list := NewGpxDataList()
	var err error
	list, err = list.FromJsonString(str)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(list.ToJsonString())
}
