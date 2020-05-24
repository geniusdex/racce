package qogs

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

// floatForValue gets a float64 value out of any numeric type
func floatForValue(value reflect.Value) float64 {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		return value.Float()
	default:
		panic(fmt.Sprintf("qogs value is not numeric (%s)", value.String()))
	}
}

// resolveCompareTo resolves an appropiate compareTo function for the given value
func resolveCompareTo(value reflect.Value) reflect.Value {
	if timeVal, ok := value.Interface().(time.Time); ok {
		return reflect.ValueOf(func(other time.Time) int {
			if timeVal.Before(other) {
				return -1
			} else if timeVal.After(other) {
				return 1
			}
			return 0
		})
	}
	compareTo := value.MethodByName("CompareTo")
	log.Printf("1 value='%v' compareTo='%v' canAddr='%v'", value.Type(), compareTo, value.CanAddr())
	if !compareTo.IsValid() && value.CanAddr() {
		log.Printf("2 value='%v' compareTo='%v'", value.Addr().Type(), value.Addr().MethodByName("CompareTo"))
		return value.Addr().MethodByName("CompareTo")
	}
	return compareTo
}

// compareValues is Compare but accepts reflection values instead.
func compareValues(first, second reflect.Value) int {
	a := elemValueFromValue(first)
	b := elemValueFromValue(second)

	if a.Kind() == reflect.String || b.Kind() == reflect.String {
		return strings.Compare(fmt.Sprint(a.Interface()), fmt.Sprint(b.Interface()))
	}

	switch a.Kind() {
	case reflect.Bool:
		if a.Bool() == b.Bool() {
			return 0
		} else if !a.Bool() {
			return -1
		}
		return 1
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		a := floatForValue(a)
		b := floatForValue(b)
		if a < b {
			return -1
		} else if a > b {
			return 1
		}
		return 0
	case reflect.Struct:
		compareTo := resolveCompareTo(a)
		if !compareTo.IsValid() {
			panic(fmt.Sprintf("qogs struct has no CompareTo method (%s)", a.String()))
		}
		if compareTo.Type().In(0).Kind() == reflect.Ptr {
			b = b.Addr()
		}
		return int(compareTo.Call([]reflect.Value{b})[0].Int())
	default:
		panic(fmt.Sprintf("qogs values are not comparable (%s) (%s)", a.String(), b.String()))
	}

}

// Compare compares the two given values using their CompareTo method. It returns
// one of the following values:
//
//    < 0   if the first argument is smaller than the second
//    == 0  if the first argument is equal to the second
//    > 0   if the first argument is larget than the second
//
// If the given arguments cannot be compared, this function panics.
//
// To make Cmp support your types, add a CompareTo method which accepts anything
// it can compare to and make it return the same values as this method does.
func Compare(first, second interface{}) int {
	return compareValues(reflect.ValueOf(first), reflect.ValueOf(second))
}
