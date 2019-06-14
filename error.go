package api

import (
	"bytes"
	"encoding/json"
	"fmt"

	api "github.com/micro/micro/api/proto"
)

// ErrHTTPCode 统一返回错误代码
var ErrHTTPCode int32 = 911

// Err Err
type Err struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error Error
func Error(rsp *api.Response, code int, format string, vals ...interface{}) error {
	rsp.StatusCode = ErrHTTPCode

	b := new(bytes.Buffer)

	json.NewEncoder(b).Encode(Err{Code: code, Message: fmt.Sprintf(format, vals...)})

	rsp.Body = b.String()

	return nil

}
