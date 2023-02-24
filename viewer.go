package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/danomagnum/go-jsonschema-generator"
)

type Viewer struct {
	Root any
}

var static_suffixes = []string{"css", "js", "ico"}

func (v *Viewer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// check for static data directory
	if strings.HasPrefix(r.URL.Path, "/static/") {
		v.ServeStatic(w, r)
		return
	}

	// check for static data types also.
	for _, sfx := range static_suffixes {
		if strings.HasSuffix(r.URL.Path, sfx) {
			v.ServeStatic(w, r)
			return
		}
	}

	// lookup actual item in config
	item, err := v.ResolvePath(r.URL.Path)
	if err != nil {
		log.Printf("problem resolving path: %v", err)
		return
	}
	switch r.Method {
	case "POST":
		// updating
		v.Edit(w, r, item)
	case "GET":
		// just viewing
		// we don't need to do anything special here.
	}
	v.View(w, r, item)
}

func (v *Viewer) Edit(w http.ResponseWriter, r *http.Request, item any) {
	jd := json.NewDecoder(r.Body)
	i := *item.(*any)
	log.Printf("PREPOST: %v", i)
	jd.Decode(&i)
	log.Printf("POST: %v", i)
}

type ViewData struct {
	Name   string
	Schema template.JS
	Data   string
}

func (v *Viewer) View(w http.ResponseWriter, r *http.Request, item any) {
	//templates, err := template.ParseGlob("./templates/*")
	//if err != nil {
	//log.Printf("Problem parsing template glob: %v", err)
	//return
	//}
	//vd := ViewData{Name: "Test"}
	//err = templates.ExecuteTemplate(w, "main.html", vd)
	//if err != nil {
	//log.Printf("problem with template. %v", err)
	//}

	var err error

	i := *item.(*any)

	d, err := json.Marshal(i)
	if err != nil {
		log.Printf("problem jsonifying: %v", err)
	}
	//w.Header().Set("content-type", "application/json")
	//w.Write(d)

	templates, err := template.ParseGlob("./templates/*")
	if err != nil {
		log.Printf("Problem parsing template glob: %v", err)
		return
	}
	s := jsonschema.Document{}
	s.Read(i)
	sb, _ := s.Marshal()
	vd := ViewData{Name: "Test", Schema: template.JS(sb), Data: string(d)}
	err = templates.ExecuteTemplate(w, "main.html", vd)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}
}

// traverse the path of struct fields, array indeces, and map keys to get to the final node we're working with.
func (v *Viewer) ResolvePath(fullpath string) (any, error) {
	path_parts := strings.Split(fullpath, "/")

	position := reflect.ValueOf(&v.Root).Elem()
	current_path := ""

	for _, path := range path_parts {
		if path == "" {
			continue
		}
		// TODO: check for arrays also.
		switch position.Kind() {
		case reflect.Struct:
			position = position.FieldByName(path)
		case reflect.Map:
			keytype := position.Type().Key()
			switch keytype.Kind() {
			case reflect.String:
				position = position.MapIndex(reflect.ValueOf(path))
			case reflect.Int:
				map_index, err := strconv.Atoi(path)
				if err != nil {
					return &reflect.Value{}, fmt.Errorf("path %s is not a valid map index", path)
				}
				position = position.MapIndex(reflect.ValueOf(map_index))
			default:
				return &reflect.Value{}, fmt.Errorf("map key type %s is not supported", keytype)
			}
		case reflect.Array, reflect.Slice:
			path_index, err := strconv.Atoi(path)
			if err != nil {
				return &reflect.Value{}, fmt.Errorf("path %s is not a valid array index", path)
			}
			if path_index >= position.Len() || path_index < 0 {
				return &reflect.Value{}, fmt.Errorf("index %d is out of bounds (expect 0..%d)", path_index, position.Len())
			}
			position = position.Index(path_index)
		default:
			return &reflect.Value{}, fmt.Errorf("path %s is not a struct. Cannot lookup %s from non-struct", current_path, path)
		}
	}

	log.Printf("can addr: %v", position.CanAddr())

	return position.Addr().Interface(), nil

}

var static_server = http.StripPrefix("/static/", http.FileServer(http.Dir("static")))

func (v *Viewer) ServeStatic(w http.ResponseWriter, r *http.Request) {
	static_server.ServeHTTP(w, r)
}
