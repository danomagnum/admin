package admin

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/gorilla/schema"
)

type Admin struct {
	Structs map[string]any
	Prefix  string
	Funcs   map[string]func()
}

func NewAdmin() *Admin {
	a := new(Admin)
	a.Structs = make(map[string]any)
	a.Prefix = "/admin"
	return a
}

var static_suffixes = []string{"css", "js", "ico"}

func (v *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := strings.TrimPrefix(r.URL.Path, v.Prefix)

	// check for static data directory
	if strings.HasPrefix(urlPath, "/static/") {
		r.URL.Path = urlPath
		v.ServeStatic(w, r)
		return
	}

	// check for static data types also.
	for _, sfx := range static_suffixes {
		if strings.HasSuffix(urlPath, sfx) {
			v.ServeStatic(w, r)
			return
		}
	}

	// lookup actual item in config
	item, err := v.ResolvePath(urlPath)
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

var decoder = schema.NewDecoder()

func (v *Admin) Edit(w http.ResponseWriter, r *http.Request, item any) {

	err := r.ParseForm()
	if err != nil {
		log.Printf("problem parsing form: %v", err)
	}

	customchange, ok := item.(Changer)
	if ok {
		newitem := reflect.New(reflect.TypeOf(item).Elem()).Interface()
		err = decoder.Decode(newitem, r.PostForm)
		if err != nil {
			log.Printf("problem decoding form: %v", err)
		}
		customchange.Change(newitem)
		return
	}

	// r.PostForm is a map of our POST form values
	err = decoder.Decode(item, r.PostForm)
	if err != nil {
		log.Printf("problem decoding form: %v", err)
	}

	n, ok := item.(Notifyable)
	if ok {
		n.Changed()
	}

}

type ViewData struct {
	Name    string
	Form    template.HTML
	Prefix  string
	Structs []string
}

func (v *Admin) View(w http.ResponseWriter, r *http.Request, item any) {
	var err error

	templates, err := template.ParseFS(templateEmbededFS, "templates/*")
	if err != nil {
		log.Printf("Problem parsing template glob: %v", err)
		return
	}

	html := StructToForm(item)

	itms := make([]string, 0, len(v.Structs))
	for k := range v.Structs {
		itms = append(itms, k)
	}
	slices.Sort(itms)
	vd := ViewData{Name: "Test", Form: html, Structs: itms, Prefix: v.Prefix}
	err = templates.ExecuteTemplate(w, "main.html", vd)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}
}

// traverse the path of struct fields, array indeces, and map keys to get to the final node we're working with.
func (v *Admin) ResolvePath(fullpath string) (any, error) {

	fullpath = strings.TrimLeft(fullpath, "/")

	parts := strings.Split(fullpath, "/")
	switch len(parts) {
	case 1:
		// this should be a direct key access.
		itm_any, ok := v.Structs[fullpath]
		if !ok {
			return nil, fmt.Errorf("key %s not found", fullpath)
		}
		_, ok = itm_any.([]any)
		if ok {
			return nil, fmt.Errorf("key %s is a list so we need an index", fullpath)
		}
		return itm_any, nil
	case 2:
		// this should be an indexed key access
		itm_any, ok := v.Structs[parts[0]]
		if !ok {
			return nil, fmt.Errorf("key %s not found", fullpath)
		}
		itm_lst, ok := itm_any.([]any)
		if !ok {
			return nil, fmt.Errorf("key %s is not a list so we need an index", fullpath)
		}
		idx, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("index %s could not be parsed: %v", parts[1], err)
		}
		if idx < 0 || idx >= len(itm_lst) {
			return nil, fmt.Errorf("index %d is out of bounds", idx)
		}
		return itm_lst[idx], nil

	default:
		return nil, fmt.Errorf("bad number of path parts: %d", len(parts))
	}
}

var static_server = http.StripPrefix("/", http.FileServer(http.FS(staticEmbededFS)))

func (v *Admin) ServeStatic(w http.ResponseWriter, r *http.Request) {
	static_server.ServeHTTP(w, r)
}

// Add an item to the admin page.
func (v *Admin) RegisterStruct(key string, val any) {
	v.Structs[key] = val
}

// Remove an item from the admin page.
func (v *Admin) UnRegisterStruct(key string) {
	delete(v.Structs, key)
}

// Add an item to the admin page.
func (v *Admin) RegisterFunc(key string, f func()) {
	v.Funcs[key] = f
}

// Remove an item from the admin page.
func (v *Admin) UnRegisterFunc(key string) {
	delete(v.Funcs, key)
}
