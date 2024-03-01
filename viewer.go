package admin

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"slices"
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
	a.Funcs = make(map[string]func())
	a.Prefix = "/admin"
	return a
}

var static_suffixes = []string{"css", "js", "ico"}

func (v *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := strings.TrimPrefix(r.URL.Path, v.Prefix)

	if urlPath == "/" {
		v.Home(w, r)
		return
	}

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

	if strings.HasPrefix(urlPath, "/edit/") {
		// lookup actual item in config
		urlPath = strings.TrimPrefix(urlPath, "/edit/")
		item, key, err := v.resolveStruct(urlPath)
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
		v.View(w, r, key, item)
	}

	if strings.HasPrefix(urlPath, "/call/") {
		// lookup actual item in config
		urlPath = strings.TrimPrefix(urlPath, "/call/")
		f, _, err := v.resolveFunc(urlPath)
		if err != nil {
			log.Printf("problem resolving path: %v", err)
			return
		}
		switch r.Method {
		case "POST":
			// updating
			v.Call(w, r, f)
		case "GET":
			// just viewing
			// we don't need to do anything special here.
			v.Call(w, r, f)
		}
		v.Home(w, r)
	}

	if strings.HasPrefix(urlPath, "/delete/") {
		// lookup actual item in config
		urlPath = strings.TrimPrefix(urlPath, "/delete/")
		item, key, err := v.resolveStruct(urlPath)
		if err != nil {
			log.Printf("problem resolving path: %v", err)
			return
		}
		switch r.Method {
		case "POST":
			// updating
			if d, ok := item.(Deleteable); ok {
				v.UnRegisterStruct(key)
				d.Delete(v)
				http.Redirect(w, r, v.Prefix, http.StatusSeeOther)
			}
		case "GET":
			// just viewing
			// we don't need to do anything special here.
			v.View(w, r, key, item)
		}
	}

}

var decoder = schema.NewDecoder()

func (v *Admin) Call(w http.ResponseWriter, r *http.Request, item func()) {
	item()
}

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
		customchange.Change(v, newitem)

		if n, ok := item.(Notifyable); ok {
			n.Changed(v)
		}

		return
	}

	// r.PostForm is a map of our POST form values
	err = decoder.Decode(item, r.PostForm)
	if err != nil {
		log.Printf("problem decoding form: %v", err)
	}

	if n, ok := item.(Notifyable); ok {
		n.Changed(v)
	}

}

type ViewData struct {
	Name       string
	Form       template.HTML
	Prefix     string
	Deleteable bool
	Structs    []string
	Funcs      []string
}

func (v *Admin) View(w http.ResponseWriter, r *http.Request, key string, item any) {
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
	fs := make([]string, 0, len(v.Funcs))
	for k := range v.Funcs {
		fs = append(fs, k)
	}
	slices.Sort(fs)
	vd := ViewData{Name: key, Form: html, Structs: itms, Prefix: v.Prefix, Funcs: fs}
	if _, ok := item.(Deleteable); ok {
		vd.Deleteable = true
	}

	err = templates.ExecuteTemplate(w, "main.html", vd)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}
}

// traverse the path of struct fields, array indeces, and map keys to get to the final node we're working with.
func (v *Admin) resolveStruct(fullpath string) (any, string, error) {

	fullpath = strings.TrimLeft(fullpath, "/")

	parts := strings.Split(fullpath, "/")
	switch len(parts) {
	case 1:
		// this should be a direct key access.
		itm_any, ok := v.Structs[parts[0]]
		if !ok {
			return nil, "", fmt.Errorf("key %s not found", parts[0])
		}
		return itm_any, parts[0], nil
	default:
		return nil, "", fmt.Errorf("bad number of path parts: %d", len(parts))
	}
}

// traverse the path of struct fields, array indeces, and map keys to get to the final node we're working with.
func (v *Admin) resolveFunc(fullpath string) (func(), string, error) {

	fullpath = strings.TrimLeft(fullpath, "/")

	parts := strings.Split(fullpath, "/")
	switch len(parts) {
	case 1:
		// this should be a direct key access.
		itm_any, ok := v.Funcs[parts[0]]
		if !ok {
			return nil, "", fmt.Errorf("key %s not found", parts[0])
		}
		return itm_any, parts[0], nil
	default:
		return nil, "", fmt.Errorf("bad number of path parts: %d", len(parts))
	}
}

var static_server = http.StripPrefix("/", http.FileServer(http.FS(staticEmbededFS)))

func (v *Admin) ServeStatic(w http.ResponseWriter, r *http.Request) {
	static_server.ServeHTTP(w, r)
}

// Add a struct to the admin page.
// val should be a pointer to a struct instance
//
// Once registered, the struct will have a link on the admin
// page to edit all its public properties.
//
// If the struct implements one of the advanced interfaces,
// additional functionality can be used.
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

func (v *Admin) Home(w http.ResponseWriter, r *http.Request) {
	var err error

	templates, err := template.ParseFS(templateEmbededFS, "templates/*")
	if err != nil {
		log.Printf("Problem parsing template glob: %v", err)
		return
	}

	itms := make([]string, 0, len(v.Structs))
	for k := range v.Structs {
		itms = append(itms, k)
	}
	slices.Sort(itms)
	fs := make([]string, 0, len(v.Funcs))
	for k := range v.Funcs {
		fs = append(fs, k)
	}
	slices.Sort(fs)
	vd := ViewData{Name: "Home", Form: "", Structs: itms, Prefix: v.Prefix, Funcs: fs}

	err = templates.ExecuteTemplate(w, "main.html", vd)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}
}
