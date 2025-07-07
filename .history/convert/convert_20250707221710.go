// Package convert implements some functions to convert data types.
package convert

import "reflect"

// ConvertToInterfaceSlice converts any slice type to []interface{}
func ConvertToInterfaceSlice(slice interface{}) []interface{} {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		// If input is not a slice, return a slice with single element
		return []interface{}{slice}
	}

	result := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}
	return result
}
