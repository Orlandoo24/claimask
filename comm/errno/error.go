package errno

import "fmt"

type BusinessError struct {
	Code    int
	Message string
	Detail  interface{}
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("[%d]%s: %v", e.Code, e.Message, e.Detail)
}

func NewError(code int, detail interface{}) error {
	return &BusinessError{
		Code:    code,
		Message: GetMsg(code),
		Detail:  detail,
	}
}
