package transform

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Note: isEmpty panics if v is nil.
func isEmpty(v interface{}) bool {
	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array, reflect.String, reflect.Chan:
		return value.Len() == 0
	}
	return false
}

type convFunc func(v interface{}) (interface{}, error)

var convStrToInt convFunc = func(v interface{}) (interface{}, error) { return strconv.ParseInt(v.(string), 10, 64) }
var convStrToFloat convFunc = func(v interface{}) (interface{}, error) { return strconv.ParseFloat(v.(string), 64) }
var convStrToBool convFunc = func(v interface{}) (interface{}, error) { return strconv.ParseBool(v.(string)) }
var convIntToFloat convFunc = func(v interface{}) (interface{}, error) { return float64(reflect.ValueOf(v).Int()), nil }
var convUintToFloat convFunc = func(v interface{}) (interface{}, error) { return float64(reflect.ValueOf(v).Uint()), nil }
var convFloatToInt convFunc = func(v interface{}) (interface{}, error) { return int64(reflect.ValueOf(v).Float()), nil }
var convToStr convFunc = func(v interface{}) (interface{}, error) { return fmt.Sprintf("%v", v), nil }

var errTypeConversionNotSupported = errors.New("type conversion not supported")

func resultTypeConversion(v interface{}, resultType resultType) (interface{}, error) {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch resultType {
		case resultTypeInt:
			return v, nil
		case resultTypeFloat:
			return convIntToFloat(v)
		case resultTypeString:
			return convToStr(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch resultType {
		case resultTypeInt:
			return v, nil
		case resultTypeFloat:
			return convUintToFloat(v)
		case resultTypeString:
			return convToStr(v)
		}
	case reflect.Float32, reflect.Float64:
		switch resultType {
		case resultTypeInt:
			return convFloatToInt(v)
		case resultTypeFloat:
			return v, nil
		case resultTypeString:
			return convToStr(v)
		}
	case reflect.Bool:
		switch resultType {
		case resultTypeBoolean:
			return v, nil
		case resultTypeString:
			return convToStr(v)
		}
	case reflect.String:
		switch resultType {
		case resultTypeInt:
			return convStrToInt(v)
		case resultTypeFloat:
			return convStrToFloat(v)
		case resultTypeBoolean:
			return convStrToBool(v)
		case resultTypeString:
			return v, nil
		}
	}
	return nil, errTypeConversionNotSupported
}

func normalizeAndSaveValue(decl *Decl, v interface{}, save func(interface{})) error {
	vv := reflect.ValueOf(v)
	if vv.Kind() == reflect.String && !decl.NoTrim {
		v = strings.TrimSpace(v.(string))
		vv = reflect.ValueOf(v)
	}
	checkToSave := func(v interface{}) {
		if v != nil && !isEmpty(v) {
			save(v)
			return
		}
		if !decl.KeepEmptyOrNull {
			return
		}
		if v == nil || vv.Kind() != reflect.String {
			save(v)
			return
		}
		save(v)
		return
	}
	if v == nil || decl.ResultType == nil {
		checkToSave(v)
		return nil
	}
	converted, err := resultTypeConversion(v, *decl.ResultType)
	if err != nil {
		return fmt.Errorf("unable to convert value '%v' to type '%s' on '%s', err: %s",
			v, *decl.ResultType, decl.fqdn, err.Error())
	}
	checkToSave(converted)
	return nil
}

func normalizeAndReturnValue(decl *Decl, v interface{}) (interface{}, error) {
	var ret interface{}
	err := normalizeAndSaveValue(decl, v, func(normalizedValue interface{}) {
		ret = normalizedValue
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}
