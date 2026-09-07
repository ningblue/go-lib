package render

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// 业务错误码分段（服务端统一；细分码可在各业务以常量追加，如 40901 租户重名）。
const (
	CodeOK                  = 0
	CodeBadRequest          = 40000
	CodeUnauthorized        = 40100
	CodeForbidden           = 40300
	CodeNotFound            = 40400
	CodeConflict            = 40900
	CodeTooMany             = 42900
	CodeInternal            = 50000
	CodeUpstreamUnavailable = 50300
)

// BizError 业务错误：Code 为上述分段（细分）码，由 RenderBizErr 渲染为统一响应。
type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string { return e.Msg }

// NewBizError 构造业务错误（格式化消息）。
func NewBizError(code int, format string, args ...interface{}) error {
	return &BizError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

// BizHTTPStatus 业务码 → HTTP 状态（分段前三位）。
func BizHTTPStatus(code int) int {
	if code == CodeOK {
		return http.StatusOK
	}
	switch code / 100 {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden,
		http.StatusNotFound, http.StatusConflict, http.StatusTooManyRequests:
		return code / 100
	case http.StatusServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// RenderBizErr 统一业务错误渲染：*BizError 按业务码分段渲染；
// 其余错误按内部错误处理（不向前端泄露实现细节）。
func RenderBizErr(c *app.RequestContext, err error) {
	var be *BizError
	if errors.As(err, &be) {
		ErrorWithCode(c, BizHTTPStatus(be.Code), be.Code, be.Msg)
		return
	}
	InternalServerError(c, "internal error")
}

// 确保 consts 引用不悬空（未来扩展分段状态时使用）。
var _ = consts.StatusOK
