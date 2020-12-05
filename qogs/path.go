// Package qogs provides primitives to query collections of nested structs.
package qogs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// evaluatePathFunction evaluates a function in a path
func evaluatePathFunction(value reflect.Value, name string, args []string) reflect.Value {
	switch name {
	case "len":
		len := pathElemValueFromValue(value, args[0]).Len()
		return reflect.ValueOf(len)
	case "tolower":
		lowered := strings.ToLower(fmt.Sprint(pathElemValueFromValue(value, args[0]).Interface()))
		return reflect.ValueOf(lowered)
	case "literal":
		literal := strings.Join(args, " ")
		return reflect.ValueOf(literal)
	default:
		panic(fmt.Sprintf("qogs invalid function call (%s %s)", name, strings.Join(args, "")))
	}
}

// pathValueFromValue is Path but accepts and returns a reflection value instead
func pathValueFromValue(value reflect.Value, path string) reflect.Value {
	value = elemValueFromValue(value)
	if path == "." {
		return value
	} else if path[0] == '.' {
		fieldName := path[1:]
		end := strings.IndexAny(fieldName, ".()") + 1 // offset by 1 to get index in path
		if end == 0 {
			return value.FieldByName(fieldName)
		}
		fieldName = path[1:end]
		return pathValueFromValue(value.FieldByName(fieldName), path[end:])
	} else if (path[0] >= '0' && path[0] <= '9') || path[0] == '-' {
		if i, err := strconv.ParseInt(path, 10, 64); err == nil {
			return reflect.ValueOf(i)
		}
		if f, err := strconv.ParseFloat(path, 64); err == nil {
			return reflect.ValueOf(f)
		}
		panic(fmt.Sprintf("qogs cannot parse numeric value '%s'", path))
	}
	entries := strings.Split(path, " ")
	return evaluatePathFunction(value, entries[0], entries[1:])
}

// pathElemValueFromValue is Path but accepts a reflection value and returns one after Elem instead
func pathElemValueFromValue(value reflect.Value, path string) reflect.Value {
	return elemValueFromValue(pathValueFromValue(value, path))
}

// pathValue is Path but returns a reflection value instead
func pathValue(data interface{}, path string) reflect.Value {
	return pathValueFromValue(elemValue(data), path)
}

// pathElemValue is Path but returns a reflection value after Elem instead
func pathElemValue(data interface{}, path string) reflect.Value {
	return elemValueFromValue(pathValue(data, path))
}

// Path looks up a value by parsing the given path.
//
// A path can be one of the following:
//
//    .<Field>          Select a field from the value
//    <fn> [<arg>...]   Call a function with optional arguments
//
// The following functions are available:
//
//   len <path>       Get the length of the array, slice, map or string
//                    identified by the given path
func Path(data interface{}, path string) interface{} {
	return pathValue(data, path).Interface()
}
