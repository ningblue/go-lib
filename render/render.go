package render

import (
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type RespJsonData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type RespListData struct {
	Total    int         `json:"total"`
	PageNum  int         `json:"page_num"`
	PageSize int         `json:"page_size"`
	List     interface{} `json:"list"`
}

// Success 响应成功消息
func Success(c *app.RequestContext) {
	c.JSON(consts.StatusOK, RespJsonData{
		Code:    CodeOK,
		Message: consts.StatusMessage(consts.StatusOK),
	})
}

// SuccessData 响应成功消息
func SuccessData(c *app.RequestContext, data interface{}) {
	c.JSON(consts.StatusOK, RespJsonData{
		Code:    CodeOK,
		Message: consts.StatusMessage(consts.StatusOK),
		Data:    data,
	})
}

// Error 响应错误消息
func Error(c *app.RequestContext, code int, message string) {
	c.JSON(code, RespJsonData{
		Code:    code,
		Message: message,
	})
}

// ErrorWithCode 业务码与 HTTP 状态分离：body.code 为业务错误码（如 40900），HTTP 状态为其分段（409）。
func ErrorWithCode(c *app.RequestContext, httpStatus, bizCode int, message string) {
	c.JSON(httpStatus, RespJsonData{
		Code:    bizCode,
		Message: message,
	})
}

// Custom 响应自定义消息
func Custom(c *app.RequestContext, code int, message string, data interface{}) {
	c.JSON(code, RespJsonData{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// BadRequest 响应错误请求消息
func BadRequest(c *app.RequestContext, message string) {
	ErrorWithCode(c, consts.StatusBadRequest, 40000, message)
}

// Unauthorized 响应未授权消息
func Unauthorized(c *app.RequestContext, message string) {
	ErrorWithCode(c, consts.StatusUnauthorized, 40100, message)
}

// Forbidden 响应禁止访问消息
func Forbidden(c *app.RequestContext, message string) {
	ErrorWithCode(c, consts.StatusForbidden, 40300, message)
}

// NotFound 响应资源未找到消息
func NotFound(c *app.RequestContext, message string) {
	ErrorWithCode(c, consts.StatusNotFound, 40400, message)
}

// Conflict 响应冲突消息（业务码 40900）
func Conflict(c *app.RequestContext, message string) {
	ErrorWithCode(c, consts.StatusConflict, 40900, message)
}

// UpstreamUnavailable 响应上游暂不可用消息（503 / 业务码 50300）
func UpstreamUnavailable(c *app.RequestContext, message string) {
	ErrorWithCode(c, consts.StatusServiceUnavailable, 50300, message)
}

// InternalServerError 响应服务器内部错误消息
func InternalServerError(c *app.RequestContext, message string) {
	ErrorWithCode(c, consts.StatusInternalServerError, 50000, message)
}

var (
	createdFailed = "创建失败"
	updateFailed  = "更新失败"
	deleteFailed  = "删除失败"
	foundFailed   = "查询失败"
)

// CreateFailed 创建失败
func CreateFailed(c *app.RequestContext, message string) {
	Error(c, consts.StatusBadRequest, createdFailed+": "+message)
}

// UpdateFailed 更新失败
func UpdateFailed(c *app.RequestContext, message string) {
	Error(c, consts.StatusBadRequest, updateFailed+": "+message)
}

// DeleteFailed 删除失败
func DeleteFailed(c *app.RequestContext, message string) {
	Error(c, consts.StatusBadRequest, deleteFailed+": "+message)
}

// FoundFailed 查询失败
func FoundFailed(c *app.RequestContext, message string) {
	Error(c, consts.StatusBadRequest, foundFailed+": "+message)
}

// 获取数组参数
func GetArrayParam(c *app.RequestContext, key string) []string {
	var result []string
	if !strings.HasSuffix(key, "[]") {
		key = key + "[]"
	}
	for _, value := range c.QueryArgs().PeekAll(key) {
		result = append(result, string(value))
	}
	return result
}

// GetArrayParamInt 获取数组参数返回int
func GetArrayParamInt(c *app.RequestContext, key string) []int {
	var result []int
	if !strings.HasSuffix(key, "[]") {
		key = key + "[]"
	}
	for _, value := range c.QueryArgs().PeekAll(key) {
		if val, err := strconv.Atoi(string(value)); err == nil {
			result = append(result, val)
		}
	}
	return result
}

// GetUsernameByUserInfo .
func GetUsernameByUserInfo(c *app.RequestContext) string {
	if info, ok := c.Get("userInfo"); ok {
		if userInfo, ok := info.(map[string]interface{}); ok {
			return userInfo["username"].(string)
		}
	}
	return ""
}

func IsAdmin(c *app.RequestContext) bool {
	if info, ok := c.Get("isAdmin"); ok {
		if isAdmin, ok := info.(bool); ok {
			return isAdmin
		}
	}
	return false
}

// GetUserID .
func GetUserID(c *app.RequestContext) string {
	if userID, ok := c.Get("userID"); ok {
		return userID.(string)
	}
	return ""
}

// GetUserRoleIDUint .
func GetUserRoleIDUint(c *app.RequestContext) uint {
	if userRoleID, ok := c.Get("roleID"); ok {
		return userRoleID.(uint)
	}
	return 0
}

// GetUserRoleIDStr .
func GetUserRoleIDStr(c *app.RequestContext) string {
	if userRoleID, ok := c.Get("roleID"); ok {
		return userRoleID.(string)
	}
	return ""
}

// GetUserRoleName .
func GetUserRoleName(c *app.RequestContext) string {
	if userRoleName, ok := c.Get("roleName"); ok {
		return userRoleName.(string)
	}
	return ""
}

