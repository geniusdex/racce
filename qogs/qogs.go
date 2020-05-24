// Package qogs provides primitives to query collections of nested structs.
package qogs

// TODO
// - Support proper paths in pathValue
// - Unit tests
// - Add Comparable interface for Sort(), model after go-linq Comparable, allow function with correct argument type
// - Implement Comparable for built-in types

import (
	"reflect"
	"sort"
)

// elemValueFromValue is Elem but accpts and returns reflection values instead.
func elemValueFromValue(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		return value.Elem()
	}
	return value
}

// elemValue is Elem but returns a reflection value instead.
func elemValue(data interface{}) reflect.Value {
	return elemValueFromValue(reflect.ValueOf(data))
}

// elemFromValue is Elem but accepts a reflection value instead.
func elemFromValue(value reflect.Value) interface{} {
	return elemValueFromValue(value).Interface()
}

// Elem optionally unwraps a pointer into the contained value. If the passed
// value is not a pointer, it returns the passed argument verbatim.
func Elem(data interface{}) interface{} {
	return elemValueFromValue(reflect.ValueOf(data)).Interface()
}

// Keys retrieves a slice of all keys in the given map.
func Keys(data interface{}) []interface{} {
	value := elemValue(data)
	result := make([]interface{}, value.Len())
	for i, key := range value.MapKeys() {
		result[i] = key.Interface()
	}
	return result
}

// Values retrieves a slice of all values in the given argument. The argument
// must be an array, slice, map or string.
func Values(data interface{}) []interface{} {
	value := elemValue(data)
	result := make([]interface{}, 0, value.Len())
	if value.Kind() == reflect.Map {
		iter := value.MapRange()
		for iter.Next() {
			result = append(result, iter.Value().Interface())
		}
	} else {
		for i := 0; i < value.Len(); i++ {
			result = append(result, value.Index(i).Interface())
		}
	}
	return result
}

// SortOn sorts the given argument on the element with the given path. The argument
// must be an array, slice, map or string. For a map, only the values are sorted and
// the keys are discarded.
func SortOn(data interface{}, path string) []interface{} {
	values := Values(data)
	sort.SliceStable(values, func(i, j int) bool {
		a := pathElemValue(values[i], path)
		b := pathElemValue(values[j], path)
		return compareValues(a, b) < 0
	})
	return values
}

// Reverse reverses the reversed values of an array, slice, map or string
func Reverse(data interface{}) []interface{} {
	values := Values(data)
	for i, j := len(values)-1, 0; i > j; i, j = i-1, j+1 {
		values[i], values[j] = values[j], values[i]
	}
	return values
}

// Contains checks if a certain value is present in an array, slice, map or string
func Contains(haystack interface{}, needle interface{}) bool {
	for _, element := range Values(haystack) {
		if Compare(element, needle) == 0 {
			return true
		}
	}
	return false
}
