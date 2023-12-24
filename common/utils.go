package common

import (
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func GetRandom64() int64 {
	n := rand.Int63()
	return n
}

func GetNowTimeString() string {
	tm := time.Now()
	return tm.Format("2006/01/02 15:04:05")
}

// 毫秒时间戳变为字符串
func TimeStampToString(t int64) string {
	tm := time.Unix(t, 0)
	//返回string
	//fmt.Printf("%-10s %-10T %s", "t", tm, tm)
	return tm.Format("2006/01/02 15:04:05")
}

func TimeStampToDateString(t int64) string {
	tm := time.Unix(t, 0)
	//返回string
	//fmt.Printf("%-10s %-10T %s", "t", tm, tm)
	return tm.Format("2006/01/02") // 20060102
}

// 获取请求中的IP
func RemoteIp(req *http.Request) string {
	var remoteAddr string
	// RemoteAddr
	remoteAddr = req.RemoteAddr
	if remoteAddr != "" {
		return remoteAddr
	}
	// ipv4
	remoteAddr = req.Header.Get("ipv4")
	if remoteAddr != "" {
		return remoteAddr
	}
	//
	remoteAddr = req.Header.Get("XForwardedFor")
	if remoteAddr != "" {
		return remoteAddr
	}
	// X-Forwarded-For
	remoteAddr = req.Header.Get("X-Forwarded-For")
	if remoteAddr != "" {
		return remoteAddr
	}
	// X-Real-Ip
	remoteAddr = req.Header.Get("X-Real-Ip")
	if remoteAddr != "" {
		return remoteAddr
	} else {
		remoteAddr = "127.0.0.1"
	}
	return remoteAddr
}

// 过滤IP
// [::1]:54706
// 192.168.1.2:54785
func ParseIp(str string) (string, string) {
	if strings.ContainsAny(str, "[]") {
		index := strings.LastIndex(str, ":")
		if index > 0 {
			return strings.Trim(str[:index], "[]"), str[index+1:]
		}
		return str, ""
	} else {
		data := strings.Split(str, ":")
		if len(data) > 1 {
			return data[0], data[1]
		}
		return str, ""
	}
}

func RemoteIpport(req *http.Request) (string, string) {
	str := RemoteIp(req)
	return ParseIp(str)
}
