package main

import "errors"

var (
	ErrNoUser        = errors.New("user name not found")
	ErrWrongPwd      = errors.New("wrong pwd")
	ErrWrongSid      = errors.New("should login first")
	ErrBadParam      = errors.New("params are bad, check manual")
	ErrNilPointer    = errors.New("null pointer data")
	ErrOldValue      = errors.New("old value")
	ErrNoValue       = errors.New("no value")
	ErrInternal      = errors.New("inernal error")
	ErrBadPermission = errors.New("permission error")
	ErrNoData        = errors.New("no data error")
)

const (
	ErrInternalCode      = 300
	ErrNoUserCode        = 301 // 用户名不对，需要注册
	ErrWrongPwdCode      = 302 //密码不对
	ErrWrongSidCode      = 303 // 重新登录，
	ErrBadParamCode      = 304 // 参数有错误
	ErrBadPermissionCode = 305
	ErrNoDataCode        = 306
)
