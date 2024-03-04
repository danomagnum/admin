package admin

import (
	"reflect"
	"strconv"
	"time"
)

func durationConverter(base time.Duration) func(string) reflect.Value {
	return func(value string) reflect.Value {
		d, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}
		}
		return reflect.ValueOf(time.Duration(d) * base)
	}
}
