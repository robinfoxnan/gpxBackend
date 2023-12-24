package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"net/http"
	"strconv"
	"strings"
	"zhituBackend/common"
	"zhituBackend/model"
	"zhituBackend/service"
)

func initUserHandlerV1(r *gin.Engine) {

	// 应用权限验证中间件到所有以 "/v1/user/" 开头的路由
	userGroup := r.Group("/v1/user")
	userGroup.Use(AuthMiddleware())
	{
		userGroup.GET("/login", WrapHttpHandler(loginHandlers))
		userGroup.GET("/regist", WrapHttpHandler(registUserHandlers)) // 注册用户，目前只支持匿名方式

		userGroup.GET("/logout", WrapHttpHandler(logoutHandlers)) // 未实现

		userGroup.GET("/searchfriends", WrapHttpHandler(searchUserHandlers))    // 搜索用户
		userGroup.GET("/addfriendreq", WrapHttpHandler(addfriendreqHandlers))   // 发送加好友请求，如果这里使用微博的关注模式，则不需要应答
		userGroup.GET("/setfriendinfo", WrapHttpHandler(setfriendinfoHandlers)) // 设置好友是否显示以及删除好友等
		userGroup.GET("/addfriendres", WrapHttpHandler(addfriendresHandlers))   // 应答好友请求，默认不需要应答，以后需要支持开启确认，，未实现
		userGroup.POST("/setbaseinfo", WrapHttpHandler(setbaseinfoHandlers))    // 设置个人基本信息

		userGroup.GET("/setrealinfo", WrapHttpHandler(setrealinfoHandlers))            // 设置个人的认证信息，包括手机，邮箱等，未实现
		userGroup.GET("/removefriend", WrapHttpHandler(removefriendHandlers))          // 删除，未实现
		userGroup.GET("/blockfriend", WrapHttpHandler(blockfriendHandlers))            // 拉黑，可以认为是设置权限的一种，未实现
		userGroup.POST("/permissionfriend", WrapHttpHandler(permissionfriendHandlers)) // 设置好友权限，未使用
		userGroup.GET("/listfriends", WrapHttpHandler(listfriendsHandlers))            // 获取好友列表

		// 添加其他用户管理的 API 路由...
	}

	//r.GET("/v1/user/regist", WrapHttpHandler(registUserHandlers))                 // 注册用户，目前只支持匿名方式
	//r.GET("/v1/user/searchfriends", WrapHttpHandler(searchUserHandlers))          // 搜索用户
	//r.GET("/v1/user/addfriendreq", WrapHttpHandler(addfriendreqHandlers))         // 发送加好友请求，如果这里使用微博的关注模式，则不需要应答
	//r.GET("/v1/user/setfriendinfo", WrapHttpHandler(setfriendinfoHandlers))       // 设置好友是否显示以及删除好友等
	//r.GET("/v1/user/addfriendres", WrapHttpHandler(addfriendresHandlers))         // 应答好友请求，默认不需要应答，以后需要支持开启确认
	//r.GET("/v1/user/setbaseinfo", WrapHttpHandler(setbaseinfoHandlers))           // 设置个人基本信息
	//r.GET("/v1/user/setrealinfo", WrapHttpHandler(setrealinfoHandlers))           // 设置个人的认证信息，包括手机，邮箱等
	//r.GET("/v1/user/removefriend", WrapHttpHandler(removefriendHandlers))         // 删除
	//r.GET("/v1/user/blockfriend", WrapHttpHandler(blockfriendHandlers))           // 拉黑，可以认为是设置权限的一种
	//r.GET("/v1/user/permissionfriend", WrapHttpHandler(permissionfriendHandlers)) // 设置好友权限
	//r.GET("/v1/user/listfriends", WrapHttpHandler(listfriendsHandlers))           // 获取好友列表

}

// AuthMiddleware 是一个简单的权限验证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在这里进行权限验证逻辑
		// 如果验证失败，可以中断请求并返回错误信息
		// 这里使用示例逻辑，你需要根据实际情况进行更复杂的权限验证

		// 在这里排除不需要验证的路由, post 在json中提交数据
		if c.FullPath() == "/v1/user/login" ||
			c.FullPath() == "/v1/user/regist" ||
			c.FullPath() == "/v1/user/setbaseinfo" ||
			c.FullPath() == "/v1/user/permissionfriend" {
			c.Next()
			return
		}

		sid := queryParamCommon(c, "sid")
		uid := queryParamCommon(c, "uid")

		if !userHasPermission(sid, uid) {

			c.JSON(http.StatusUnauthorized, gin.H{
				"state":  "fail",
				"code":   model.ErrWrongSidCode,
				"detail": model.ErrWrongSid.Error(),
			})
			c.Abort() // 中断请求
			return
		}

		// 如果权限验证通过，继续处理下一个中间件或路由处理函数
		c.Next()
	}
}

// userHasPermission 是一个示例权限验证函数，你需要根据实际情况实现
func userHasPermission(sid, uid string) bool {
	// 在实际应用中，你需要根据传递的 token 或其他凭证来验证用户权限
	// 这里只是一个示例，你可能需要调用认证服务或数据库来检查权限
	if len(sid) == 0 {
		return false
	}

	ok := service.CheckUserSession(sid, uid)
	if !ok {
		return false
	}
	return true
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
		sess, err := service.LoginUserSession(uid, sid)
		if err != nil {
			temp := fmt.Sprintf(`{"state": "fail", "code":%d, "detail":"%s" `, model.ErrWrongSidCode, model.ErrWrongSid.Error())
			w.Write([]byte(temp))
			return
		} else {
			data, _ := json.Marshal(sess)
			temp := fmt.Sprintf(`{"state": "ok", "session":%s }`, string(data))
			w.Write([]byte(temp))
			return
		}
	} else if rtype == "1" {
		sess, err := service.LoginUserBase(uid, pwd)
		if err != nil {
			temp := fmt.Sprintf(`{"state": "fail", "code":%d, "detail":“%s”}`, model.ErrWrongPwdCode, err.Error())
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
	ip, _ := common.RemoteIpport(r)
	//fmt.Println(ip)

	r.ParseForm()
	values := r.URL.Query()
	rtype := values.Get("type")
	name := values.Get("username")
	pwd := values.Get("pwd")
	fmt.Printf("%s, %s \n", name, pwd)
	userinfo, ret := service.RegistUser(rtype, name, pwd, ip)
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
	user, ret := service.SearchUser(uid, fid)
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

	fmt.Println("addfriendreqHandlers", values)
	err := service.AddUserFollow(fid, uid)
	if err != nil {
		temp := fmt.Sprintf(`{"state": "fail", "code": %d, "deitail":"%s"}`, model.ErrInternalCode, "add user follower err")
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
	param := values.Get("param") //

	fmt.Println("setfriendinfoHandlers", values)
	err := service.SetUserFollowSimple(fid, uid, param)
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
			user, err := model.UserInfoFormBytes(body)
			if err == nil {

				uid := strconv.FormatInt(user.Id, 10)
				sid := strconv.FormatInt(user.Session, 10)
				ok := service.CheckUserSession(sid, uid)
				if !ok {
					temp := fmt.Sprintf(`{"state": "fail", "code": %d, "detail":"%s"}`, model.ErrWrongSidCode, model.ErrWrongSid.Error())
					w.Write([]byte(temp))
					return
				}

				err = service.SetUserBaseInfo(user)
				if err == nil {
					temp := fmt.Sprintf(`{"state": "ok", "detail":"user json updated"}`)
					w.Write([]byte(temp))
				} else {
					temp := fmt.Sprintf(`{"state": "fail", "detail":"update err:%s"}`, err.Error())
					w.Write([]byte(temp))
				}
			} else {
				temp := fmt.Sprintf(`{"state": "fail", "detail":"user json err:%s"}`, err.Error())
				w.Write([]byte(temp))
			}

		}
	} else {
		temp := fmt.Sprintf(`{"state": "fail", "deitail":"not finish"}`)
		w.Write([]byte(temp))
	}

}

// 请求绑定手机或者邮件的验证，未实现
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
			var param model.FriendPermissinParam
			err := json.Unmarshal(body, &param)
			if err == nil {

				//fmt.Println(param)
				uid := strconv.FormatInt(param.Id, 10)
				sid := strconv.FormatInt(param.Sid, 10)
				fid := strconv.FormatInt(param.Fid, 10)
				ok := service.CheckUserSession(sid, uid)
				if !ok {
					temp := fmt.Sprintf(`{"state": "fail", "code": %d, "detail":“%s”}`, model.ErrWrongSidCode, model.ErrWrongSid.Error())
					w.Write([]byte(temp))
					return
				}

				err = service.SetUserFunPermission(uid, sid, fid, param.Opt)
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
	//sid := values.Get("sid")
	//ok := service.CheckUserSession(sid, uid)
	//if !ok {
	//	temp := fmt.Sprintf(`{"state": "fail", "code": %d, "detail":“%s”}`, model.ErrWrongSidCode, model.ErrWrongSid.Error())
	//	w.Write([]byte(temp))
	//	return
	//}
	gtype := values.Get("type")
	if gtype == "1" {
		lst, _ := service.GetUserFollowList(uid, 0, 0)
		str, _ := model.FriendListToString(lst)
		temp := fmt.Sprintf(`{"state": "ok", "count":%d, "list": %s}`, len(lst), str)
		w.Write([]byte(temp))
		return
	} else {
		lst, _ := service.GetUserFunList(uid, 0, 0)
		str, _ := json.Marshal(lst)
		temp := fmt.Sprintf(`{"state": "ok", "count":%d, "list": %s}`, len(lst), str)
		w.Write([]byte(temp))
		return
	}

	temp := fmt.Sprintf(`{"state": "fail", "deitail":“not finish”}`)
	w.Write([]byte(temp))
}
