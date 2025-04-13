// Package response 封装响应结构
package response

import (
	"claimask/comm/constant"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ERROR   = -1
	SUCCESS = 0
)

// Response 响应结构体
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Msg     string      `json:"msg"`
	TraceID string      `json:"traceId"`
}

// Result trace Id
func Result(code int, data interface{}, msg string, c *gin.Context) {
	c.Set(constant.ERROR_CODE, code)
	traceID := c.GetString(constant.TRACKING_ID)
	// 开始时间
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Data:    data,
		Msg:     msg,
		TraceID: traceID,
	})
}

// Ok result
func Ok(c *gin.Context) {
	Result(SUCCESS, map[string]interface{}{}, "操作成功", c)
}

// OkWithMessage result
func OkWithMessage(c *gin.Context, message string) {
	Result(SUCCESS, map[string]interface{}{}, message, c)
}

// OkWithData result
func OkWithData(c *gin.Context, data interface{}) {
	Result(SUCCESS, data, "操作成功", c)
}

// OkWithDetailed result
func OkWithDetailed(c *gin.Context, data interface{}, message string) {
	Result(SUCCESS, data, message, c)
}

// Fail result
func Fail(c *gin.Context, code int) {
	Result(code, map[string]interface{}{}, "操作失败", c)
}

// FailWithMessage result
func FailWithMessage(c *gin.Context, code int, message string) {
	Result(code, map[string]interface{}{}, message, c)
}
