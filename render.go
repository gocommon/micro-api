package api

import (
	"bytes"
	"encoding/json"

	api "github.com/micro/micro/api/proto"
)

// JSON JSON
func JSON(rsp *api.Response, code int32, i interface{}) error {
	rsp.StatusCode = code

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(i)

	if err != nil {
		return Error(rsp, 2345, "NewEncoder err ")
	}
	rsp.Body = b.String()

	return nil
}
