// Package errno 定义错误码，错误码统一放在这边集中管理，不允许在各自包中定义错误码
package errno

import "fmt"

// ErrNo 定义异常码对象
type ErrNo struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// Error 返回错误对象
func (e *ErrNo) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Msg)
}

var (
	SUCCESS = ErrNo{0, "success"}

	ERR_BAD_REQUEST     = &ErrNo{400, "bad request error"}
	ERR_UNAUTHORIZED    = &ErrNo{401, "unauthorized error"}
	ERR_FORBIDDEN       = &ErrNo{403, "forbidden error"}
	ERR_CONFLICT        = &ErrNo{409, "conflict error"}
	ERR_INTERNAL_SERVER = &ErrNo{500, "internal server error"}
)
