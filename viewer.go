package admin

import (
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/schema"
)

type Admin struct {
	Structs  map[string]any
	Funcs    map[string]func()
	timeBase time.Duration
	Prefix   string
	decoder  *schema.Decoder
}

func NewAdmin(options ...func(*Admin)) *Admin {
	a := new(Admin)
	a.Structs = make(map[string]any)
	a.Funcs = make(map[string]func())
	for _, o := range options {
		o(a)
	}

	a.decoder = schema.NewDecoder()
	if a.timeBase != 0 {
		a.decoder.RegisterConverter(time.Duration(0), durationConverter(a.timeBase))
	}

	return a
}

func (v *Admin) Mux(prefix string) http.Handler {
	v.Prefix = prefix
	var static_server = http.StripPrefix("/", http.FileServer(http.FS(staticEmbededFS)))

	mux := http.NewServeMux()
	mux.HandleFunc("/", v.Home)
	mux.Handle("/static/", static_server)
	mux.HandleFunc("POST /edit/{key}", v.edit)
	mux.HandleFunc("GET /edit/{key}", v.view)
	mux.HandleFunc("/call/{key}", v.call)
	mux.HandleFunc("POST /delete/{key}", v.delete) // this should be DELETE probably, but browsers don't support it in a form natively without JS.
	return http.StripPrefix(prefix, mux)
}

func (v *Admin) delete(w http.ResponseWriter, r *http.Request) {
	urlPath := r.PathValue("key")
	if urlPath == "" {
		slog.Warn("No key provided")
		v.Home(w, r)
		return
	}
	slog.Info("Got delete request", "alarmid", urlPath)
	item, key, err := v.resolveStruct(urlPath)
	if err != nil {
		log.Printf("problem resolving path: %v", err)
		return
	}
	v.UnRegisterStruct(key)
	if d, ok := item.(Deleteable); ok {
		d.AdminDelete(v)
	}
	http.Redirect(w, r, v.Prefix, http.StatusSeeOther)
}

func (v *Admin) call(w http.ResponseWriter, r *http.Request) {

	urlPath := r.PathValue("key")
	if urlPath == "" {
		slog.Warn("No key provided")
		v.Home(w, r)
		return
	}
	f, _, err := v.resolveFunc(urlPath)
	if err != nil {
		log.Printf("problem resolving path: %v", err)
		return
	}
	f()

	v.Home(w, r)

}

func (v *Admin) edit(w http.ResponseWriter, r *http.Request) {

	urlPath := r.PathValue("key")
	if urlPath == "" {
		slog.Warn("No key provided")
		v.Home(w, r)
		return
	}

	item, _, err := v.resolveStruct(urlPath)
	if err != nil {
		log.Printf("problem resolving path: %v", err)
		return
	}
	err = r.ParseForm()
	if err != nil {
		log.Printf("problem parsing form: %v", err)
	}

	customchange, ok := item.(Changer)
	if ok {
		newitem := reflect.New(reflect.TypeOf(item).Elem()).Interface()
		err = v.decoder.Decode(newitem, r.PostForm)
		if err != nil {
			log.Printf("problem decoding form: %v", err)
		}
		customchange.AdminChange(v, newitem)

		if n, ok := item.(Notifyable); ok {
			n.AdminChanged(v)
		}

		v.view(w, r)
		return
	}

	// r.PostForm is a map of our POST form values
	err = v.decoder.Decode(item, r.PostForm)
	if err != nil {
		log.Printf("problem decoding form: %v", err)
	}

	if n, ok := item.(Notifyable); ok {
		n.AdminChanged(v)
	}

	v.view(w, r)
}

type StructDescriptor struct {
	Name    string
	Display string
}

type ViewData struct {
	Name       string
	Form       template.HTML
	Deleteable bool
	Structs    []StructDescriptor
	Funcs      []string
	Status     string
	Prefix     string
}

func (v *Admin) view(w http.ResponseWriter, r *http.Request) {

	urlPath := r.PathValue("key")
	if urlPath == "" {
		slog.Warn("No key provided")
		v.Home(w, r)
		return
	}

	item, key, err := v.resolveStruct(urlPath)
	if err != nil {
		log.Printf("problem resolving path: %v", err)
		return
	}

	templates, err := template.ParseFS(templateEmbededFS, "templates/*")
	if err != nil {
		log.Printf("Problem parsing template glob: %v", err)
		return
	}

	html := StructToForm(item, v.timeBase)

	itms := make([]StructDescriptor, 0)
	//slices.Sort(itms)
	fs := make([]string, 0, len(v.Funcs))
	for k := range v.Funcs {
		fs = append(fs, k)
	}
	slices.Sort(fs)

	status := ""
	sts_itm, ok := item.(StatusIndicating)
	if ok {
		status = sts_itm.AdminStatus()
	}
	vd := ViewData{Name: key, Form: html, Structs: itms, Prefix: v.Prefix, Funcs: fs, Status: status}
	if _, ok := item.(Deleteable); ok {
		vd.Deleteable = true
	}

	err = templates.ExecuteTemplate(w, "view.html", vd)
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

	itms := make([]StructDescriptor, 0, len(v.Structs))
	for k := range v.Structs {
		desc := StructDescriptor{Name: k}
		cd, ok := v.Structs[k].(CustomDisplay)
		if ok {
			desc.Display = cd.AdminDisplay()
		} else {
			desc.Display = k
		}

		itms = append(itms, desc)
	}
	slices.SortFunc(itms, func(i, j StructDescriptor) int {
		if i.Name == "New Alarm" {
			return -1
		}
		if j.Name == "New Alarm" {
			return 1
		}
		return strings.Compare(i.Name, j.Name)
	})
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
