package gowebstructapi

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type FieldDescriptor struct {
	Name  string
	Value any
	Kind  reflect.Kind
}

func GetNameToFieldMap(model any) []FieldDescriptor {
	t := reflect.TypeOf(model).Elem()
	v := reflect.ValueOf(model).Elem()
	nameToDataPointerMap := make([]FieldDescriptor, v.NumField())
	for i := range nameToDataPointerMap {
		name := t.Field(i).Name
		typ := t.Field(i).Type.Kind()
		structValue := reflect.Indirect(v).FieldByName(name)
		nameToDataPointerMap[i] = FieldDescriptor{Name: name, Value: structValue.Interface(), Kind: typ}
	}
	return nameToDataPointerMap
}

func StructToForm(model any) string {
	m := GetNameToFieldMap(model)

	s := strings.Builder{}
	for _, v := range m {
		switch x := v.Value.(type) {
		case bool:
			if x {
				s.WriteString(fmt.Sprintf("<input type='checkbox' name='%s' checked>\n", v.Name))
			} else {
				s.WriteString(fmt.Sprintf("<input type='checkbox' name='%s'>\n", v.Name))
			}
		case int, string, byte, int16, uint16, int32, uint32, int64, uint64, float32, float64:
			s.WriteString(fmt.Sprintf("<input type='text' name='%s' value='%v'>\n", v.Name, x))
		default:
			s.WriteString(fmt.Sprintf("!! Problem with %s (%T)", v.Name, x))
		}

	}

	return s.String()

}

func RespToStruct[T any](r *http.Request) (*T, error) {
	//m := GetNameToFieldMap(model)

	var realt T
	t := &realt
	err := r.ParseForm()
	if err != nil {
		return t, fmt.Errorf("problem parsing form: %w", err)
	}

	rv := reflect.ValueOf(t).Elem()
	for key, val := range r.Form {
		if len(val) != 1 {
			return t, fmt.Errorf("only support single values per field so far but %s had %d", key, len(val))
		}

		fv := rv.FieldByName(key)

		switch fv.Kind() {
		case reflect.Bool:
			switch val[0] {
			case "true", "TRUE", "True", "1", "on":
				fv.SetBool(true)
			case "false", "FALSE", "False", "0", "off":
				fv.SetBool(false)
			default:
				return t, fmt.Errorf("could not parse %s as bool for %s", val[0], key)
			}
		case reflect.Int:
			v, err := strconv.ParseInt(val[0], 10, 64)
			if err != nil {
				return t, fmt.Errorf("could not parse int return value for %s: %w", key, err)
			}
			fv.SetInt(v)
		case reflect.Float32, reflect.Float64:
			v, err := strconv.ParseFloat(val[0], 64)
			if err != nil {
				return t, fmt.Errorf("could not parse float return value for %s: %w", key, err)
			}
			fv.SetFloat(v)
		case reflect.String:
			fv.SetString(val[0])
		}

	}

	return t, nil

}
