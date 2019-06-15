package errors

import (
	"errors"
	"testing"
)

func Test_parse(t *testing.T) {
	err := &Error{
		Code:    404,
		Message: "文件不存在",
		OrgErr:  errors.New("file not found"),
	}

	str := "404:文件不存在【err】file not found"

	if str != err.Error() {
		t.Fail()
	}

	err2 := Parse(str)
	if !ErrorEqual(err2, err) {
		t.Logf("%#v != %#v", err2.Error(), err.Error())
		t.Fail()
	}
}
