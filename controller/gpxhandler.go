package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"zhituBackend/db"
	"zhituBackend/model"
	"zhituBackend/service"
)

var chGpxData chan *model.GpxData = nil
var chanList chan *model.GpxDataArray = nil

// 关闭这个，会保证协程停止
func StopGpxChannal() {
	if chGpxData != nil {
		close(chGpxData)
		chGpxData = nil
	}

	if chanList != nil {
		close(chanList)
		chanList = nil
	}
}

func initGpxHandlerV1(r *gin.Engine) {

	chGpxData, chanList = service.StartGpxStoreWorker(db.RedisCli)
	// 这里主要还是将点的管理加入进去
	r.POST("/v1/gpx/updatepoint", WrapHttpHandler(gpxHandlers))      // 单个点上报
	r.POST("/v1/gpx/updatepoints", WrapHttpHandler(gpxListHandlers)) // 一组点上报

	r.GET("/v1/gpx/position", WrapHttpHandler(getLastPtHandlers))  // 获取某个好友最后的位置
	r.POST("/v1/gpx/position", WrapHttpHandler(getLastPtHandlers)) // 获取某个好友最后的位置
	r.POST("/v1/gpx/track", WrapHttpHandler(getTrackHandlers))     // 获取轨迹

}

// /////////////////////////////////////////////////////////////////
// 解析gpx数据并写入管道
// 优先解析body的json数据，其次是解析URL中的数据
func gpxHandlers(w http.ResponseWriter, r *http.Request) {

	//fmt.Println(r.Method)
	//fmt.Println("req from: " + r.RemoteAddr)
	//fmt.Println("headers:")
	//if len(r.Header) > 0 {
	//	for k, v := range r.Header {
	//		fmt.Printf("%s=%s\n", k, v[0])
	//	}
	//}
	//tm1 := time.Now().UnixMicro()

	var gpx *model.GpxData
	var err error
	// post method, should parse body
	if strings.EqualFold("Post", r.Method) {
		// 获取请求报文的内容长度
		len := r.ContentLength
		if len > 0 {
			body := make([]byte, len)
			r.Body.Read(body)
			//fmt.Println("body=" + string(body))

			gpx, err = model.GpxDataFromJson(string(body))
			if err != nil {
				//w.WriteHeader(500)
				fmt.Fprintln(w, `{"state": "fail", "detail": "parse json meet error"}`)
				//logger.Error("upload gpx point err= " + err.Error())
				return
			}
		}
	}

	// 如果是 get方法
	if gpx == nil {
		gpx = &model.GpxData{}
	}

	str := r.FormValue("id")
	if len(str) > 0 {
		gpx.Uid = str
	}

	str = r.FormValue("lat")
	if len(str) > 0 {
		gpx.Lat, _ = strconv.ParseFloat(str, 64)
	}
	str = r.FormValue("lon")
	if len(str) > 0 {
		gpx.Lon, _ = strconv.ParseFloat(str, 64)
	}
	str = r.FormValue("ele")
	if len(str) > 0 {
		gpx.Ele, _ = strconv.ParseFloat(str, 64)
	}

	str = r.FormValue("speed")
	if len(str) > 0 {
		gpx.Speed, _ = strconv.ParseFloat(str, 64)
	}

	str = r.FormValue("tm")
	if len(str) > 0 {
		gpx.Tm, _ = strconv.ParseInt(str, 10, 64)
	}

	// 解析后无错误的情况下，写入管道，等待持久化
	if chGpxData != nil {
		chGpxData <- gpx
	} else {
		fmt.Println("error: channel is nil\n")
	}

	//fmt.Println(gpx.ToJsonString())

	w.WriteHeader(200)
	fmt.Fprintln(w, `{"state": "ok" }`)

	//tm2 := time.Now().UnixMicro()
	//delta := tm2 - tm1
	//fmt.Printf("cost: %d ms\n", delta/1000.0)
}

// 当长久时间未上线后，会发送一组数据，可以使用此批量接收接口
func gpxListHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	if !strings.EqualFold("Post", r.Method) {
		temp := fmt.Sprintf(`{"state": "fail", "deitail":“not get finish”}`)
		w.Write([]byte(temp))
	}

	var err error
	var list *model.GpxDataArray = model.NewGpxDataList()
	// 获取请求报文的内容长度
	len := r.ContentLength
	if len > 0 {
		body := make([]byte, len)
		r.Body.Read(body)
		//fmt.Println("body=" + string(body))

		list, err = list.FromJsonString(string(body))
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, "gpxList has err %s", err)
			return
		}
		// post list to channel
		chanList <- list
	}

	w.WriteHeader(200)
	fmt.Fprintln(w, `{"state": "ok" }`)
}

// 使用post提交的请求
func getLastPtHandlersPost(w http.ResponseWriter, r *http.Request) {

	var err error
	param := model.QueryParamPoint{}
	// 获取请求报文的内容长度
	len := r.ContentLength
	if len > 0 {
		body := make([]byte, len)
		r.Body.Read(body)
		//fmt.Println("body=" + string(body))

		err = param.QueryParamPointFromJson(string(body))
		if err != nil {
			w.WriteHeader(200)
			//fmt.Fprintln(w, "param has err: %s", err.Error())
			temp := fmt.Sprintf(`{"state": "fail", "deitail":“%s”, "code": %d}`, err.Error(), model.ErrBadParamCode)
			fmt.Fprintln(w, temp)
			return
		}
	}

	//// 先查用户
	//ok := CheckUserSession(param.Sid, param.Uid)
	//if !ok {
	//	temp := fmt.Sprintf(`{"state": "fail", "code": %d, "deitail":"%s"}`, model.ErrWrongSidCode, model.ErrWrongSid.Error())
	//	fmt.Fprintln(w, temp)
	//	return
	//}

	// 查看是否在对方好友列表中
	// 先在对方的粉丝中查询权限
	ret := service.CheckUserPermission(param.Uid, param.Fid)
	if !ret {
		temp := fmt.Sprintf(`{"state": "fail", "deitail":"permission err", "code": %d}`, model.ErrBadPermissionCode)
		fmt.Fprintln(w, temp)
		return
	}

	// 直接查询
	gpx, err := db.RedisCli.FindLastGpx(param.Fid)
	if gpx == nil {
		w.WriteHeader(200)
		//fmt.Fprintln(w, "param has err: %s", err.Error())
		temp := fmt.Sprintf(`{"state": "fail", "deitail":"%s", "code": %d}`, err.Error(), model.ErrNoDataCode)
		fmt.Fprintln(w, temp)
		return
	}

	gpxJson, err := gpx.ToJsonString()
	if err != nil {
		w.WriteHeader(200)
		//fmt.Fprintln(w, "param has err: %s", err.Error())

		temp := fmt.Sprintf(`{"state": "fail", "deitail":"%s", "code": %d}`, err.Error(), model.ErrNoDataCode)
		fmt.Fprintln(w, temp)
		return
	}

	w.WriteHeader(200)
	//data := `{"state": "ok", "pt":` + gpxJson + `}`
	// fmt.Fprintln(w, data)
	temp := fmt.Sprintf(`{"state": "ok", "pt": %s }`, gpxJson)
	fmt.Fprintln(w, temp)
	return
}

func getLastPtHandlersGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	r.ParseForm()
	values := r.URL.Query()
	uid := values.Get("uid")
	//	sid := values.Get("sid")
	fid := values.Get("fid")

	// 查看是否在对方好友列表中
	// 先在对方的粉丝中查询权限
	ret := service.CheckUserPermission(uid, fid)
	if !ret {
		temp := fmt.Sprintf(`{"state": "fail", "deitail":"permission err", "code": %d}`, model.ErrBadPermissionCode)
		fmt.Fprintln(w, temp)
		return
	}

	// 直接查询
	gpx, err := db.RedisCli.FindLastGpx(fid)
	if gpx == nil {
		w.WriteHeader(200)
		//fmt.Fprintln(w, "param has err: %s", err.Error())
		temp := fmt.Sprintf(`{"state": "fail", "deitail":"%s", "code": %d}`, "", model.ErrNoDataCode)
		fmt.Fprintln(w, temp)
		return
	}

	gpxJson, err := gpx.ToJsonString()
	if err != nil {
		w.WriteHeader(200)
		//fmt.Fprintln(w, "param has err: %s", err.Error())

		temp := fmt.Sprintf(`{"state": "fail", "deitail":"%s", "code": %d}`, err.Error(), model.ErrNoDataCode)
		fmt.Fprintln(w, temp)
		return
	}

	w.WriteHeader(200)
	//data := `{"state": "ok", "pt":` + gpxJson + `}`
	// fmt.Fprintln(w, data)
	temp := fmt.Sprintf(`{"state": "ok", "pt": %s }`, gpxJson)
	fmt.Fprintln(w, temp)
	return
}

// 查询某人最后位置
func getLastPtHandlers(w http.ResponseWriter, r *http.Request) {
	if strings.EqualFold("Post", r.Method) {
		getLastPtHandlersPost(w, r)
	} else {
		getLastPtHandlersGet(w, r)
	}
}

// 查询某人一段时间轨迹
func getTrackHandlers(w http.ResponseWriter, r *http.Request) {
	if !strings.EqualFold("Post", r.Method) {
		w.WriteHeader(500)
		fmt.Fprintln(w, "positon should use Post method")
		return
	}

	var err error
	param := model.QueryParamTrack{}
	// 获取请求报文的内容长度
	len := r.ContentLength
	if len > 0 {
		body := make([]byte, len)
		r.Body.Read(body)
		//fmt.Println("body=" + string(body))

		err = param.QueryParamTrackFromJson(string(body))
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, "param has err: %s", err)
			return
		}
	}

	// 直接查询
	gpxList, err := db.RedisCli.FindGpxTrack(&param)
	if gpxList == nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "find track err: %s", err)
		return
	}

	gpxJson, err := gpxList.ToJsonString()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, ": %s", err)
		return
	}

	w.WriteHeader(200)
	data := `{"state": "OK", "ptList":` + gpxJson + `}`
	fmt.Fprintln(w, data)
}
