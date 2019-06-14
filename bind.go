package api

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	api "github.com/micro/micro/api/proto"
)

type (

	// BindUnmarshaler is the interface used to wrap the UnmarshalParam method.
	BindUnmarshaler interface {
		// UnmarshalParam decodes and assigns a value from an form or query param.
		UnmarshalParam(param string) error
	}
)

// Bind Bind
func Bind(req *api.Request, i interface{}) error {
	contentLength := req.Header["Content-Length"]

	req.Method = strings.ToUpper(req.Method)

	if contentLength == nil {
		if req.Method == "GET" || req.Method == "DELETE" {
			if err := bindData(i, req.Get, "json"); err != nil {
				return err
			}
			return nil
		}
		return errors.New("Request body can't be empty")
	}

	var ctype string
	contentType, has := req.Header["Content-Type"]
	if has {
		ctype = contentType.Values[0]
	}

	bodyRd := strings.NewReader(req.Body)

	switch {
	case strings.HasPrefix(ctype, "application/json"):
		if err := json.NewDecoder(bodyRd).Decode(i); err != nil {
			if ute, ok := err.(*json.UnmarshalTypeError); ok {
				return fmt.Errorf("Unmarshal type error: expected=%v, got=%v, offset=%v", ute.Type, ute.Value, ute.Offset)
			} else if se, ok := err.(*json.SyntaxError); ok {
				return fmt.Errorf("Syntax error: offset=%v, error=%v", se.Offset, se.Error())
			} else {
				return err
			}
		}
	case strings.HasPrefix(ctype, "application/xml"), strings.HasPrefix(ctype, "text/xml"):
		if err := xml.NewDecoder(bodyRd).Decode(i); err != nil {
			if ute, ok := err.(*xml.UnsupportedTypeError); ok {
				return fmt.Errorf("Unsupported type error: type=%v, error=%v", ute.Type, ute.Error())
			} else if se, ok := err.(*xml.SyntaxError); ok {
				return fmt.Errorf("Syntax error: line=%v, error=%v", se.Line, se.Error())
			} else {
				return err
			}
		}
	case strings.HasPrefix(ctype, "application/x-www-form-urlencoded"), strings.HasPrefix(ctype, "multipart/form-data"):
		if err := bindData(i, req.Post, "json"); err != nil {
			return err
		}
	default:
		return errors.New(http.StatusText(http.StatusUnsupportedMediaType))
	}
	return nil
}

func bindData(ptr interface{}, data map[string]*api.Pair, tag string) error {
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()

	if typ.Kind() != reflect.Struct {
		return errors.New("Binding element must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}
		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get(tag)

		if inputFieldName == "" {
			inputFieldName = typeField.Name
			// If tag is nil, we inspect if the field is a struct.
			if _, ok := bindUnmarshaler(structField); !ok && structFieldKind == reflect.Struct {
				err := bindData(structField.Addr().Interface(), data, tag)
				if err != nil {
					return err
				}
				continue
			}
		}

		// 有逗号，取逗号前数据
		if idx := strings.Index(inputFieldName, ","); idx > -1 {
			inputFieldName = string(inputFieldName[0:idx])
		}

		inputValue, exists := data[inputFieldName]
		if !exists {
			continue
		}

		// Call this first, in case we're dealing with an alias to an array type
		if ok, err := unmarshalField(typeField.Type.Kind(), inputValue.Values[0], structField); ok {
			if err != nil {
				return err
			}

			continue
		}

		numElems := len(inputValue.Values)
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for j := 0; j < numElems; j++ {
				if err := setWithProperType(sliceOf, inputValue.Values[j], slice.Index(j)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else {
			if err := setWithProperType(typeField.Type.Kind(), inputValue.Values[0], structField); err != nil {
				return err
			}
		}
	}

	return nil
}

func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	// But also call it here, in case we're dealing with an array of BindUnmarshalers
	if ok, err := unmarshalField(valueKind, val, structField); ok {
		return err
	}

	switch valueKind {
	case reflect.Ptr:
		return setWithProperType(structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("unknown type")
	}
	return nil
}

func unmarshalField(valueKind reflect.Kind, val string, field reflect.Value) (bool, error) {
	switch valueKind {
	case reflect.Ptr:
		return unmarshalFieldPtr(val, field)
	default:
		return unmarshalFieldNonPtr(val, field)
	}
}

// bindUnmarshaler attempts to unmarshal a reflect.Value into a BindUnmarshaler
func bindUnmarshaler(field reflect.Value) (BindUnmarshaler, bool) {
	ptr := reflect.New(field.Type())
	if ptr.CanInterface() {
		iface := ptr.Interface()
		if unmarshaler, ok := iface.(BindUnmarshaler); ok {
			return unmarshaler, ok
		}
	}
	return nil, false
}

func unmarshalFieldNonPtr(value string, field reflect.Value) (bool, error) {
	if unmarshaler, ok := bindUnmarshaler(field); ok {
		err := unmarshaler.UnmarshalParam(value)
		field.Set(reflect.ValueOf(unmarshaler).Elem())
		return true, err
	}
	return false, nil
}

func unmarshalFieldPtr(value string, field reflect.Value) (bool, error) {
	if field.IsNil() {
		// Initialize the pointer to a nil value
		field.Set(reflect.New(field.Type().Elem()))
	}
	return unmarshalFieldNonPtr(value, field.Elem())
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
