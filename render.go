package api

import (
	"bytes"
	"encoding/json"

	"github.com/gocommon/micro-api/errors"

	api "github.com/micro/micro/api/proto"
)

// JSON JSON
func JSON(rsp *api.Response, code int32, i interface{}) error {
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(i)

	if err != nil {
		return errors.New(err, 2345, "json encode response")
	}

	rsp.StatusCode = code
	rsp.Body = b.String()

	return nil
}
