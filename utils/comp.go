package utils

import (
	"encoding/json"
	"errors"
	"gorm.io/datatypes"
	"reflect"
)

// CompareStruct 对比两个结构体是否相同
func CompareStruct(oldData, newData interface{}, fields []string) (isSame bool, err error) {
	if oldData != nil && reflect.TypeOf(oldData).Kind() != reflect.Ptr {
		return false, errors.New("传入的结构体必须为指针类型")
	}

	if newData != nil && reflect.TypeOf(newData).Kind() != reflect.Ptr {
		return false, errors.New("传入的结构体必须为指针类型")
	}

	if oldData == nil || newData == nil {
		return false, nil
	}

	return compStruct(oldData, newData, fields)
}

func compStruct(oldData, newData interface{}, fields []string) (bool, error) {
	oldRefElem := reflect.ValueOf(oldData).Elem()
	newRefElem := reflect.ValueOf(newData).Elem()

	for _, field := range fields {
		if oldRefElem.Kind() == reflect.Ptr {
			return false, errors.New("不支持指针获取字段,请检查传值类型")
		}

		oldValue := oldRefElem.FieldByName(field)
		newValue := newRefElem.FieldByName(field)

		if !oldValue.IsValid() || !newValue.IsValid() {
			continue
		}

		if oldValue.Kind() != newValue.Kind() {
			return false, errors.New("对比字段类型不统一")
		}

		// 获取字段类型
		fieldType := oldValue.Type()

		// 处理特殊类型
		if isJSONType(fieldType) {
			equal, err := compareJSONValues(oldValue.Interface(), newValue.Interface())
			if err != nil {
				return false, err
			}
			if !equal {
				return false, nil
			}
			continue
		}

		// 其他类型直接比较
		if !reflect.DeepEqual(oldValue.Interface(), newValue.Interface()) {
			return false, nil
		}
	}

	return true, nil
}

// isJSONType 判断是否是JSON类型
func isJSONType(t reflect.Type) bool {
	// 检查是否是 datatypes.JSON 检查是否是 json.RawMessage
	if t == reflect.TypeOf(datatypes.JSON{}) || t == reflect.TypeOf(json.RawMessage{}) {
		return true
	}

	return false
}

// compareJSONValues 比较两个JSON值是否相等
func compareJSONValues(a, b interface{}) (bool, error) {
	var jsonBytesA, jsonBytesB []byte
	var err error

	// 根据类型获取JSON字节
	switch v := a.(type) {
	case datatypes.JSON:
		jsonBytesA = v
	case json.RawMessage:
		jsonBytesA = v
	default:
		return false, errors.New("unsupported JSON type")
	}

	switch v := b.(type) {
	case datatypes.JSON:
		jsonBytesB = v
	case json.RawMessage:
		jsonBytesB = v
	default:
		return false, errors.New("unsupported JSON type")
	}

	// 解析JSON
	var objA, objB interface{}
	if err = json.Unmarshal(jsonBytesA, &objA); err != nil {
		return false, err
	}
	if err = json.Unmarshal(jsonBytesB, &objB); err != nil {
		return false, err
	}

	// 规范化比较
	normalizedA, err := json.Marshal(objA)
	if err != nil {
		return false, err
	}
	normalizedB, err := json.Marshal(objB)
	if err != nil {
		return false, err
	}

	return string(normalizedA) == string(normalizedB), nil
}
