package errors

import (
	"bytes"
	"encoding/json"
	"fmt"

	api "github.com/micro/micro/api/proto"
)

// ErrHTTPCode 统一返回错误代码
var ErrHTTPCode int32 = 911

// Error Error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	OrgErr  error  `json:"-"` // 原始error, 不需要返回给客户端，用于打印错误日志
}

func (p *Error) Error() string {
	return fmt.Sprintf("%d-%s:%v", p.Code, p.Message, p.OrgErr)
}

// JSON JSON
func (p *Error) JSON(rsp *api.Response) error {
	rsp.StatusCode = ErrHTTPCode

	b := new(bytes.Buffer)

	json.NewEncoder(b).Encode(p)

	rsp.Body = b.String()

	return nil
}

// New New
func New(rsp *api.Response, err error, code int, format string, vals ...interface{}) error {
	return &Error{
		OrgErr:  err,
		Code:    code,
		Message: fmt.Sprintf(format, vals...),
	}
}
