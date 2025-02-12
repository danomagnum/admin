package admin

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type FieldDescriptor struct {
	Name   string
	Descr  string
	Hidden bool
	Value  any
	Kind   reflect.Kind
}

func GetNameToFieldMap(model any) []FieldDescriptor {
	t := reflect.TypeOf(model).Elem()
	v := reflect.ValueOf(model).Elem()
	nameToDataPointerMap := make([]FieldDescriptor, 0, v.NumField())
	for i := range v.NumField() {
		if !t.Field(i).IsExported() {
			continue
		}
		name := t.Field(i).Name
		typ := t.Field(i).Type.Kind()
		structValue := reflect.Indirect(v).FieldByName(name)
		descr := t.Field(i).Tag.Get("descr")
		HiddenTag := t.Field(i).Tag.Get("adminHidden")
		hidden := false
		if HiddenTag == "true" {
			hidden = true
		}

		nameToDataPointerMap = append(nameToDataPointerMap, FieldDescriptor{Name: name, Value: structValue.Interface(), Kind: typ, Descr: descr, Hidden: hidden})
	}
	return nameToDataPointerMap
}

var formHiddenTemplate = template.Must(template.New("formHidden").Parse(`<input type="hidden" name="{{.Name}}" value="{{.Value}}">`))
var formBoolTemplate = template.Must(template.New("formBool").Parse(`
<fieldset>
	<label for='{{.Name}}' title='{{.Descr}}'>
		{{.Name}}
	</label>
	<input type='hidden' name='{{.Name}}' value='false'>
	<input type="checkbox" name="{{.Name}}" {{if .Value}}checked{{end}}>
	{{ if .Descr }}
		<div class='formComment'>
			{{.Descr}}
		</div>
	{{ end }}
</fieldset>`))
var formStringTemplate = template.Must(template.New("formString").Parse(`
<fieldset>
	<label for='{{.Name}}' title='{{.Descr}}'>
		{{.Name}}
	</label>
	<input type="text" name="{{.Name}}" value='{{.Value}}'>
	{{ if .Descr }}
		<div class='formComment'>{{.Descr}}</div>
	{{ end }}
</fieldset>`))
var formNumberTemplate = template.Must(template.New("formNumber").Parse(`
<fieldset>
	<label for='{{.Name}}' title='{{.Descr}}'>
		{{.Name}}
	</label>
	<input type="number" name="{{.Name}}" value='{{.Value}}'>
	{{ if .Descr }}
		<div class='formComment'>
			{{.Descr}}
		</div>
	{{ end }}
</fieldset>`))
var formDateTimeTemplate = template.Must(template.New("formDateTime").Parse(`
<fieldset>
	<label for='{{.Name}}' title='{{.Descr}}'>
		{{.Name}}
	</label>
	<input type="datetime-local" name="{{.Name}}" value='{{.Value}}'>
	{{ if .Descr }}
		<div class='formComment'>
			{{.Descr}}
		</div>
	{{ end }}
</fieldset>`))
var formUnknownTemplate = template.Must(template.New("formUnknown").Parse(`
<fieldset>
    <label for='{{.Name}}' title='{{.Descr}}'>
		{{.Name}} (Uknown Type!!) 
    </label>
    <input type="text" name="{{.Name}}" value='{{.Value}}'>
    {{ if .Descr }}
		<div class='formComment'>
			{{.Descr}}
		</div>
    {{ end }}
</fieldset>`))

func StructToForm(model any, timebase time.Duration) template.HTML {
	m := GetNameToFieldMap(model)
	var err error

	s := strings.Builder{}
	for _, v := range m {
		if v.Hidden {
			err = formHiddenTemplate.Execute(&s, v)
			if err != nil {
				slog.Error("Error executing template", "err", err)
			}
			//s.WriteString(fmt.Sprintf("<input type='hidden' name='%s' value='false'>\n", v.Name))
			continue
		}
		switch x := v.Value.(type) {
		case bool:
			// checkboxes don't send their values on form updates if they're not checked.  So gorilla/schema
			// will not update them to false.  This hidden field takes care of that.
			err = formBoolTemplate.Execute(&s, v)
			if err != nil {
				slog.Error("Error executing template", "err", err)
			}
		case string:
			err = formStringTemplate.Execute(&s, v)
			if err != nil {
				slog.Error("Error executing template", "err", err)
			}
		case int, byte, int16, uint16, int32, uint32, int64, uint64, float32, float64:
			err = formNumberTemplate.Execute(&s, v)
			if err != nil {
				slog.Error("Error executing template", "err", err)
			}
		case time.Duration:
			v.Value = x / timebase
			err = formNumberTemplate.Execute(&s, v)
			if err != nil {
				slog.Error("Error executing template", "err", err)
			}
		case time.Time:
			err = formDateTimeTemplate.Execute(&s, v)
			if err != nil {
				slog.Error("Error executing template", "err", err)
			}
		default:
			err = formUnknownTemplate.Execute(&s, v)
			if err != nil {
				slog.Error("Error executing template", "err", err)
			}
		}
	}

	// because we've already ran this string through the template engine, we can safely ignore the linter warning here
	//
	// #nosec G203
	return template.HTML(s.String())

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
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v, err := strconv.ParseUint(val[0], 10, 64)
			if err != nil {
				return t, fmt.Errorf("could not parse int return value for %s: %w", key, err)
			}
			fv.SetUint(v)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
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
		case reflect.Invalid, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.Struct, reflect.UnsafePointer:
			return t, fmt.Errorf("unsupported type %s for %s", fv.Kind(), key)
		default:
			return t, fmt.Errorf("unsupported type %s for %s", fv.Kind(), key)
		}

	}

	return t, nil

}
