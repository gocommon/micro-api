package errors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	api "github.com/micro/micro/api/proto"
)

// ErrHTTPCode 统一返回错误代码
var ErrHTTPCode int32 = 911

// OrgErrSeparation OrgErrSeparation
const OrgErrSeparation = "【err】"

// CodeSeparation CodeSeparation
const CodeSeparation = ":"

// Error Error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	OrgErr  error  `json:"-"` // 原始error, 不需要返回给客户端，用于打印错误日志
}

func (p *Error) Error() string {
	return fmt.Sprintf("%d%s%s%s%v", p.Code, CodeSeparation, p.Message, OrgErrSeparation, p.OrgErr)
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
func New(err error, code int, format string, vals ...interface{}) error {
	return &Error{
		OrgErr:  err,
		Code:    code,
		Message: fmt.Sprintf(format, vals...),
	}
}

// Parse Parse
func Parse(str string) *Error {

	e := &Error{}

	if len(str) == 0 {
		return e
	}

	a1 := strings.Split(str, OrgErrSeparation)
	if len(a1) > 1 {
		e.OrgErr = errors.New(strings.Join(a1[1:], OrgErrSeparation))
	}

	if idx := strings.Index(a1[0], CodeSeparation); idx > -1 {
		var err error
		e.Code, err = strconv.Atoi(string(a1[0][:idx]))
		if err != nil {
			e.Message = a1[0]
		} else {
			e.Message = string(a1[0][idx+1:])
		}

	} else {
		e.Message = a1[0]
	}

	return e
}

// ErrorEqual ErrorEqual
func ErrorEqual(e1, e2 *Error) bool {
	return e1.Code == e2.Code && e1.Message == e2.Message && e1.OrgErr.Error() == e2.OrgErr.Error()
}

// IsAPIError IsAPIError是不是
func IsAPIError(str string) bool {
	if strings.Index(str, OrgErrSeparation) < 0 {
		return false
	}

	idx := strings.Index(str, CodeSeparation)
	if idx < 0 {
		return false
	}

	_, err := strconv.Atoi(str[:idx])
	if err != nil {
		return false
	}

	return true
}
