package ngin

import (
	"time"

	"github.com/niexqc/nlibs/ntools"
)

const (
	RespCode_OK           = 0
	RespCode_Valid_Err    = 1000 // 验证错误
	RespCode_RunTime_Err  = 2000 // 运行时异常
	RespCode_RunTime_Err2 = 3000 // 捕获上游异常，转运行时异常
	RespCode_UnLogin      = 9000 // 登录过期
	RespCode_UnKnown_Err  = 9999 // 其他错误
)

// EmptyObj ...
type EmptyObj struct {
}

// BaseReq ...
type BaseReq struct {
	//
}

// BaseReqPage ...
type BaseReqPage struct {
	BaseReq
	// 页码，从1开始
	PageNo ReqVoInt `json:"pageNo" zhdesc:"页码" binding:"required,gte=1"`
	// 页码，每页大小，必须大于1
	PageSize ReqVoInt `json:"pageSize" zhdesc:"每页大小" binding:"required,gte=1"`
}

func NewReqPage(pageNo, pageSize int) BaseReqPage {
	return BaseReqPage{
		PageNo:   NewReqVoInt(pageNo),
		PageSize: NewReqVoInt(pageSize),
	}
}

// BaseResp 基础响应
type BaseResp struct {
	// 响应码
	Code int `json:"code"`
	// 响应消息
	Msg string `json:"msg"`
	// 列表时的数据总量
	Count int64 `json:"count"`
	// 列表时的数据总页
	PageCount int64 `json:"pageCount"`
	// 返回成功时可能包含警告信息
	Warn string `json:"warn"`
	// 服务器当前时间
	ServerTime string `json:"serverTime"`
	// 每个接口返回的数据类型不同，在接口文档中具体指定
	Data any `json:"data" swaggerignore:"true" `
	// 每个接口返回的数据类型不同，在接口文档中具体指定
	ExtData any `json:"extData" swaggerignore:"true" `
}

// BaseReq ...
type NiexqGinHeaderVo struct {
	UserAgent     string `json:"userAgent" zhdesc:"用户浏览器"`
	ContentType   string `json:"contentType" zhdesc:"请求的类型" `
	ContentLength int64  `json:"contentLength" zhdesc:"请求体的长度"`

	UserToken string `json:"userToken" zhdesc:"用户访问凭证" header:"user-token"`

	AppType string `json:"clientType" zhdesc:"客户端类型" `
	AppVer  string `json:"clientVer" zhdesc:"客户端版本" `

	ClientTime string `json:"clientTime" zhdesc:"客户端时间" `
	OneceStr   string `json:"oneceStr" zhdesc:"随机字符串" `
	VisitSrc   string `json:"visitSrc" zhdesc:"请求发起的源" `
	VisitSign  string `json:"visitSign" zhdesc:"本次请求的签名" `

	VisitTar string `json:"VisitTar" zhdesc:"请求访问的目标" `
	UserIp   string `json:"clientIp" zhdesc:"客户端Ip" `
	ReqBody  []byte `json:"-" zhdesc:"请求原始的Body" `
}

func emptyObj() *BaseResp {
	instance := new(BaseResp)
	instance.ServerTime = ntools.Time2Str(time.Now())
	instance.Data = EmptyObj{}
	instance.ExtData = EmptyObj{}
	return instance
}

// NewErrBaseResp ...
func NewErrBaseResp(msg string) *BaseResp {
	instance := emptyObj()
	instance.Code = RespCode_UnKnown_Err
	instance.Msg = msg
	return instance
}

// NewOkBaseResp ...
func NewOkBaseResp(data interface{}) *BaseResp {
	instance := emptyObj()
	instance.Data = data
	instance.Code = 0
	return instance
}

// NewNoBaseResp ...
func NewNoBaseResp(code int, msg string) *BaseResp {
	instance := emptyObj()
	instance.Code = code
	instance.Msg = msg
	return instance
}

// NewNoBaseResp ...
func VoIfErr(vo any, err error) *BaseResp {
	if nil != err {
		return NewErrBaseResp(err.Error())
	} else {
		return NewOkBaseResp(vo)
	}
}
