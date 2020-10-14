package transform

import (
	"fmt"
	"reflect"
	"strconv"
)

func isEmptyOrNull(v interface{}) bool {
	if v == nil {
		return true
	}
	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.String:
		return v.(string) == ""
	case reflect.Slice, reflect.Map:
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

func resultTypeConversion(v interface{}, resultType ResultType) (interface{}, error) {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch resultType {
		case ResultTypeInt:
			return v, nil
		case ResultTypeFloat:
			return convIntToFloat(v)
		case ResultTypeString:
			return convToStr(v)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch resultType {
		case ResultTypeInt:
			return v, nil
		case ResultTypeFloat:
			return convUintToFloat(v)
		case ResultTypeString:
			return convToStr(v)
		}
	case reflect.Float32, reflect.Float64:
		switch resultType {
		case ResultTypeInt:
			return convFloatToInt(v)
		case ResultTypeFloat:
			return v, nil
		case ResultTypeString:
			return convToStr(v)
		}
	case reflect.Bool:
		switch resultType {
		case ResultTypeBoolean:
			return v, nil
		case ResultTypeString:
			return convToStr(v)
		}
	case reflect.String:
		switch resultType {
		case ResultTypeInt:
			return convStrToInt(v)
		case ResultTypeFloat:
			return convStrToFloat(v)
		case ResultTypeBoolean:
			return convStrToBool(v)
		case ResultTypeString:
			return v, nil
		}
	}
	return nil, fmt.Errorf("unable to convert value (%v) of type '%T' to result_type '%s'", v, v, resultType)
}

func normalizeAndSaveValue(decl *Decl, v interface{}, save func(interface{})) error {
	if isEmptyOrNull(v) {
		if decl.KeepEmptyOrNull {
			save(nil)
		}
		return nil
	}
	if decl.ResultType == nil || *decl.ResultType == ResultTypeObject {
		save(v)
		return nil
	}
	v, err := resultTypeConversion(v, *decl.ResultType)
	if err != nil {
		return err
	}
	// The reason we don't do another isEmptyOrNull/decl.KeepEmptyOrNull check
	// here is because nothing comes out of resultTypeConversion will be null/empty.
	save(v)
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
