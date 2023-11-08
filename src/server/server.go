package main

import (
	"fmt"
	json "github.com/json-iterator/go"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

var chGpxData chan *GpxData
var chanList chan *GpxDataArray

func startHttpServer() {
	http.HandleFunc("/", indexHandler)

	// 第一阶段，主要是注册好友，添加好友等所有好友的相关操作
	http.HandleFunc("/v1/user/login", loginHandlers)   //
	http.HandleFunc("/v1/user/logout", logoutHandlers) //

	http.HandleFunc("/v1/user/regist", registUserHandlers)                 // 注册用户，目前只支持匿名方式
	http.HandleFunc("/v1/user/searchfriends", searchUserHandlers)          // 搜索用户
	http.HandleFunc("/v1/user/addfriendreq", addfriendreqHandlers)         // 发送加好友请求，如果这里使用微博的关注模式，则不需要应答
	http.HandleFunc("/v1/user/setfriendinfo", setfriendinfoHandlers)       // 设置好友是否显示以及删除好友等
	http.HandleFunc("/v1/user/addfriendres", addfriendresHandlers)         // 应答好友请求，默认不需要应答，以后需要支持开启确认
	http.HandleFunc("/v1/user/setbaseinfo", setbaseinfoHandlers)           // 设置个人基本信息
	http.HandleFunc("/v1/user/setrealinfo", setrealinfoHandlers)           // 设置个人的认证信息，包括手机，邮箱等
	http.HandleFunc("/v1/user/removefriend", removefriendHandlers)         // 删除
	http.HandleFunc("/v1/user/blockfriend", blockfriendHandlers)           // 拉黑，可以认为是设置权限的一种
	http.HandleFunc("/v1/user/permissionfriend", permissionfriendHandlers) // 设置好友权限
	http.HandleFunc("/v1/user/listfriends", listfriendsHandlers)           // 获取好友列表

	// 这里主要还是将点的管理加入进去
	http.HandleFunc("/v1/gpx/updatepoint", gpxHandlers)      // 单个点上报
	http.HandleFunc("/v1/gpx/updatepoints", gpxListHandlers) // 一组点上报
	http.HandleFunc("/v1/gpx/position", getLastPtHandlers)   // 获取某个好友最后的位置
	http.HandleFunc("/v1/gpx/track", getTrackHandlers)       // 获取轨迹

	// 第2阶段，共享信息的群组管理

	// 第3阶段，用户P2P聊天，群聊

	// 第4阶段，语音和视频功能

	// 第5阶段，添加扩展功能

	port := config.Server.Port
	logger.Info("server start:", zap.Int("port", port))
	// os.Getenv("PORT")
	//log.Printf("Defaulting to port %s", port)

	//log.Printf("Listening on port %s", port)
	//log.Printf("Open http://localhost:%s in the browser", port)
	fmt.Printf("Open http://localhost:%d in the browser", port)
	addr := fmt.Sprintf("%s:%d", config.Server.Host, port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

}

// 默认的解析函数
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, err := fmt.Fprint(w, "Hello, Welcome to Bird2Fish gpx World!")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// 0是使用sid尝试登录，1是使用用户名密码登录，2是手机验证码，3是邮箱
func loginHandlers(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("Content-Type", "application-json; charset=utf-8")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	r.ParseForm()
	values := r.URL.Query()
	rtype := values.Get("type")
	uid := values.Get("uid")
	sid := values.Get("sid")
	pwd := values.Get("pwd")

	if rtype == "0" {
		sess, err := LoginUserSession(uid, sid)
		if err != nil {
			temp := fmt.Sprintf(`{"state": "fail", "code":%d, "detail":"%s" `, ErrWrongSidCode, ErrWrongSid.Error())
			w.Write([]byte(temp))
			return
		} else {
			data, _ := json.Marshal(sess)
			temp := fmt.Sprintf(`{"state": "ok", "session":%s }`, string(data))
			w.Write([]byte(temp))
			return
		}
	} else if rtype == "1" {
		sess, err := LoginUserBase(uid, pwd)
		if err != nil {
			temp := fmt.Sprintf(`{"state": "fail", "code":%d, "detail":“%s”}`, ErrWrongPwdCode, ErrWrongPwd.Error())
			w.Write([]byte(temp))
			return
		} else {
			data, _ := json.Marshal(sess)
			temp := fmt.Sprintf(`{"state": "ok", "session":%s }`, string(data))

			//w.Header().Set("Content-Length", strconv.Itoa(len(temp)))
			//w.Write([]byte(temp))
			fmt.Fprintln(w, temp)
			fmt.Println(temp)
			return
		}
	}

	temp := fmt.Sprintf(`{"state": "fail", "deitail":“not finish”}`)
	w.Write([]byte(temp))
}
func logoutHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")
	temp := fmt.Sprintf(`{"state": "fail", "deitail":“not finish”}`)
	w.Write([]byte(temp))
}

// 注册函数
// ?type=1&code=1234&pwd=123456&username=aaaa
func registUserHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")
	//fmt.Println(r.Method)
	//if r.Method == "POST" {

	//body := r.FormValue("body")
	ip, _ := RemoteIpport(r)
	//fmt.Println(ip)

	r.ParseForm()
	values := r.URL.Query()
	rtype := values.Get("type")
	name := values.Get("username")
	pwd := values.Get("pwd")
	fmt.Printf("%s, %s \n", name, pwd)
	userinfo, ret := RegistUser(rtype, name, pwd, ip)
	//ret = true
	if ret != nil {
		temp1 := fmt.Sprintf(`{"state": "fail", "des":"%s"}`, ret.Error())
		w.Write([]byte(temp1))
		return
	}

	userInfoJson, _ := userinfo.UserInfoToString()
	temp := fmt.Sprintf(`{"state": "ok", "user":%s}`, userInfoJson)
	w.Write([]byte(temp))

}

// /v1/user/searchfriends?id=1003
func searchUserHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")
	//fmt.Println(r.Method)
	//if r.Method == "POST" {

	//body := r.FormValue("body")

	r.ParseForm()
	values := r.URL.Query()
	uid := values.Get("uid")
	fid := values.Get("fid")
	user, ret := SearchUser(uid, fid)
	if ret != nil {
		temp1 := fmt.Sprintf(`{"state": "fail", "des":"%s"}`, ret.Error())
		w.Write([]byte(temp1))
		return
	}

	userInfoJson, _ := user.UserInfoToString()
	fmt.Println("search user:", userInfoJson)
	temp := fmt.Sprintf(`{"state": "ok", "user":%s}`, userInfoJson)
	w.Write([]byte(temp))

}

// 申请添加好友
func addfriendreqHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	r.ParseForm()
	values := r.URL.Query()
	uid := values.Get("uid")
	fid := values.Get("fid")
	sid := values.Get("sid")
	ok := CheckUserSession(sid, uid)
	if !ok {

		temp := fmt.Sprintf(`{"state": "fail", "code": %d, "detail":“%s”}`, ErrWrongSidCode, ErrWrongSid.Error())
		w.Write([]byte(temp))
		return
	}

	fmt.Println("addfriendreqHandlers", values)
	err := AddUserFollow(fid, uid)
	if err != nil {
		temp := fmt.Sprintf(`{"state": "fail", "code": %d, "deitail":“%s”}`, ErrInternalCode, err.Error())
		w.Write([]byte(temp))
		return
	}

	temp := fmt.Sprintf(`{"state": "ok", "detail": "add friend ok"  }`)
	w.Write([]byte(temp))
}

// 设置好友的基本信息，包括取消关注
func setfriendinfoHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	r.ParseForm()
	values := r.URL.Query()
	uid := values.Get("uid")
	fid := values.Get("fid")
	sid := values.Get("sid")
	param := values.Get("param")
	ok := CheckUserSession(sid, uid)
	if !ok {

		temp := fmt.Sprintf(`{"state": "fail", "code": %d, "detail":“%s”}`, ErrWrongSidCode, ErrWrongSid.Error())
		w.Write([]byte(temp))
		return
	}

	fmt.Println("setfriendinfoHandlers", values)
	err := SetUserFollowSimple(fid, uid, param)
	if err == nil {

		temp := fmt.Sprintf(`{"state": "ok", "detail": "set friend info ok"  }`)
		w.Write([]byte(temp))
		return
	}

	temp := fmt.Sprintf(`{"state": "fail", "deitail":"update err:%s"}`, err.Error())
	w.Write([]byte(temp))

}

// 设置应答结果
func addfriendresHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")
	temp := fmt.Sprintf(`{"state": "fail", "deitail":“not finish”}`)
	w.Write([]byte(temp))
}

func setbaseinfoHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	// post method, should parse body
	if strings.EqualFold("Post", r.Method) {
		// 获取请求报文的内容长度
		len := r.ContentLength
		if len > 0 {
			body := make([]byte, len)
			r.Body.Read(body)
			//fmt.Println("body=" + string(body))
			user, err := UserInfoFormBytes(body)
			if err == nil {

				uid := strconv.FormatInt(user.Id, 10)
				sid := strconv.FormatInt(user.Session, 10)
				ok := CheckUserSession(sid, uid)
				if !ok {
					temp := fmt.Sprintf(`{"state": "fail", "code": %d, "detail":“%s”}`, ErrWrongSidCode, ErrWrongSid.Error())
					w.Write([]byte(temp))
					return
				}

				err = SetUserBaseInfo(user)
				if err == nil {
					temp := fmt.Sprintf(`{"state": "ok", "deitail":"user json updated"}`)
					w.Write([]byte(temp))
				} else {
					temp := fmt.Sprintf(`{"state": "fail", "deitail":"update err:%s"}`, err.Error())
					w.Write([]byte(temp))
				}
			} else {
				temp := fmt.Sprintf(`{"state": "fail", "deitail":"user json err:%s"}`, err.Error())
				w.Write([]byte(temp))
			}

		}
	} else {
		temp := fmt.Sprintf(`{"state": "fail", "deitail":"not finish"}`)
		w.Write([]byte(temp))
	}

}

func setrealinfoHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")
	temp := fmt.Sprintf(`{"state": "fail", "deitail":“not finish”}`)
	w.Write([]byte(temp))
}

func removefriendHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")
	temp := fmt.Sprintf(`{"state": "fail", "deitail":“not finish”}`)
	w.Write([]byte(temp))
}

func blockfriendHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")
	temp := fmt.Sprintf(`{"state": "fail", "deitail":“not finish”}`)
	w.Write([]byte(temp))
}

// 设置粉丝的权限
func permissionfriendHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	// post method, should parse body
	if strings.EqualFold("Post", r.Method) {
		// 获取请求报文的内容长度
		len := r.ContentLength
		if len > 0 {
			body := make([]byte, len)
			r.Body.Read(body)
			//fmt.Println("body=" + string(body))
			var param FriendPermissinParam
			err := json.Unmarshal(body, &param)
			if err == nil {

				//fmt.Println(param)
				uid := strconv.FormatInt(param.Id, 10)
				sid := strconv.FormatInt(param.Sid, 10)
				fid := strconv.FormatInt(param.Fid, 10)
				ok := CheckUserSession(sid, uid)
				if !ok {
					temp := fmt.Sprintf(`{"state": "fail", "code": %d, "detail":“%s”}`, ErrWrongSidCode, ErrWrongSid.Error())
					w.Write([]byte(temp))
					return
				}

				err = SetUserFunPermission(uid, sid, fid, param.Opt)
				if err == nil {
					temp := fmt.Sprintf(`{"state": "ok", "deitail":"friend permission updated"}`)
					w.Write([]byte(temp))
				} else {
					temp := fmt.Sprintf(`{"state": "fail", "deitail":"update err:%s"}`, err.Error())
					w.Write([]byte(temp))
				}
			} else {
				temp := fmt.Sprintf(`{"state": "fail", "deitail":"friend permission param json err:%s"}`, err.Error())
				w.Write([]byte(temp))
			}

		}
	} else {
		temp := fmt.Sprintf(`{"state": "fail", "deitail":"not finish"}`)
		w.Write([]byte(temp))
	}
}

// 获取关注的列表或者粉丝列表
func listfriendsHandlers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application-json")

	r.ParseForm()
	values := r.URL.Query()
	uid := values.Get("uid")
	sid := values.Get("sid")
	ok := CheckUserSession(sid, uid)
	if !ok {
		temp := fmt.Sprintf(`{"state": "fail", "code": %d, "detail":“%s”}`, ErrWrongSidCode, ErrWrongSid.Error())
		w.Write([]byte(temp))
		return
	}
	gtype := values.Get("type")
	if gtype == "1" {
		lst, _ := GetUserFollowList(uid, 0, 0)
		str, _ := FriendListToString(lst)
		temp := fmt.Sprintf(`{"state": "ok", "count":%d, "list": %s}`, len(lst), str)
		w.Write([]byte(temp))
		return
	} else {
		lst, _ := GetUserFunList(uid, 0, 0)
		str, _ := json.Marshal(lst)
		temp := fmt.Sprintf(`{"state": "ok", "count":%d, "list": %s}`, len(lst), str)
		w.Write([]byte(temp))
		return
	}

	temp := fmt.Sprintf(`{"state": "fail", "deitail":“not finish”}`)
	w.Write([]byte(temp))
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

	var gpx *GpxData
	var err error
	// post method, should parse body
	if strings.EqualFold("Post", r.Method) {
		// 获取请求报文的内容长度
		len := r.ContentLength
		if len > 0 {
			body := make([]byte, len)
			r.Body.Read(body)
			//fmt.Println("body=" + string(body))

			gpx, err = GpxDataFromJson(string(body))
			if err != nil {
				//w.WriteHeader(500)
				fmt.Fprintln(w, `"state": "fail", "detail": "parse json meet error: %s"`, err.Error())
				return
			}
		}
	}

	// 如果是 get方法
	if gpx == nil {
		gpx = &GpxData{}
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
	var list *GpxDataArray = NewGpxDataList()
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
	param := QueryParamPoint{}
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
			temp := fmt.Sprintf(`{"state": "fail", "deitail":“%s”, "code": %d}`, err.Error(), ErrBadParamCode)
			fmt.Fprintln(w, temp)
			return
		}
	}

	// 先查用户
	ok := CheckUserSession(param.Sid, param.Uid)
	if !ok {
		temp := fmt.Sprintf(`{"state": "fail", "code": %d, "detail":“%s”}`, ErrWrongSidCode, ErrWrongSid.Error())
		fmt.Fprintln(w, temp)
		return
	}

	// 查看是否在对方好友列表中
	// 先在对方的粉丝中查询权限
	ret := CheckUserPermission(param.Uid, param.Fid)
	if !ret {
		temp := fmt.Sprintf(`{"state": "fail", "deitail":"permission err", "code": %d}`, ErrBadPermissionCode)
		fmt.Fprintln(w, temp)
		return
	}

	// 直接查询
	gpx, err := redisCli.FindLastGpx(param.Fid)
	if gpx == nil {
		w.WriteHeader(200)
		//fmt.Fprintln(w, "param has err: %s", err.Error())
		temp := fmt.Sprintf(`{"state": "fail", "deitail":“%s”, "code": %d}`, err.Error(), ErrNoDataCode)
		fmt.Fprintln(w, temp)
		return
	}

	gpxJson, err := gpx.ToJsonString()
	if err != nil {
		w.WriteHeader(200)
		//fmt.Fprintln(w, "param has err: %s", err.Error())

		temp := fmt.Sprintf(`{"state": "fail", "deitail":“%s”, "code": %d}`, err.Error(), ErrNoDataCode)
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
	sid := values.Get("sid")
	fid := values.Get("fid")
	// 先查一下登录情况
	ok := CheckUserSession(sid, uid)
	if !ok {
		temp := fmt.Sprintf(`{"state": "fail", "code": %d, "detail":“%s”}`, ErrWrongSidCode, ErrWrongSid.Error())
		w.Write([]byte(temp))
		return
	}

	// 查看是否在对方好友列表中
	// 先在对方的粉丝中查询权限
	ret := CheckUserPermission(uid, fid)
	if !ret {
		temp := fmt.Sprintf(`{"state": "fail", "deitail":"permission err", "code": %d}`, ErrBadPermissionCode)
		fmt.Fprintln(w, temp)
		return
	}

	// 直接查询
	gpx, err := redisCli.FindLastGpx(fid)
	if gpx == nil {
		w.WriteHeader(200)
		//fmt.Fprintln(w, "param has err: %s", err.Error())
		temp := fmt.Sprintf(`{"state": "fail", "deitail":“%s”, "code": %d}`, err.Error(), ErrNoDataCode)
		fmt.Fprintln(w, temp)
		return
	}

	gpxJson, err := gpx.ToJsonString()
	if err != nil {
		w.WriteHeader(200)
		//fmt.Fprintln(w, "param has err: %s", err.Error())

		temp := fmt.Sprintf(`{"state": "fail", "deitail":“%s”, "code": %d}`, err.Error(), ErrNoDataCode)
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
	param := QueryParamTrack{}
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
	gpxList, err := redisCli.FindGpxTrack(&param)
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
