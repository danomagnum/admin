package gowebstructapi

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/schema"
	"github.com/joncalhoun/form"
)

type Admin struct {
	Data map[string]any
}

func NewAdmin() *Admin {
	a := new(Admin)
	a.Data = make(map[string]any)
	return a
}

var static_suffixes = []string{"css", "js", "ico"}

func (v *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

var decoder = schema.NewDecoder()

func (v *Admin) Edit(w http.ResponseWriter, r *http.Request, item any) {

	err := r.ParseForm()
	if err != nil {
		log.Printf("problem parsing form: %v", err)
	}

	// r.PostForm is a map of our POST form values
	err = decoder.Decode(item, r.PostForm)
	if err != nil {
		log.Printf("problem decoding form: %v", err)
	}

}

type ViewData struct {
	Name string
	Form template.HTML
}

func (v *Admin) View(w http.ResponseWriter, r *http.Request, item any) {
	var err error

	templates, err := template.ParseFS(templateEmbededFS, "templates/*")
	if err != nil {
		log.Printf("Problem parsing template glob: %v", err)
		return
	}

	tpl := template.Must(template.New("").Parse(`
   	 <input type='{{.Type}}' name='{{.Name}}' {{with .Value}}value='{{.}}'{{end}}>
   `))
	fb := form.Builder{InputTemplate: tpl}
	html, err := fb.Inputs(item)
	if err != nil {
		log.Printf("problem creating form: %v", err)
	}
	_ = html

	html = StructToForm(item)

	vd := ViewData{Name: "Test", Form: html}
	err = templates.ExecuteTemplate(w, "main.html", vd)
	if err != nil {
		log.Printf("problem with template. %v", err)
	}
}

// traverse the path of struct fields, array indeces, and map keys to get to the final node we're working with.
func (v *Admin) ResolvePath(fullpath string) (any, error) {

	return v.Data["Test"], nil

}

var static_server = http.StripPrefix("/", http.FileServer(http.FS(staticEmbededFS)))

func (v *Admin) ServeStatic(w http.ResponseWriter, r *http.Request) {
	static_server.ServeHTTP(w, r)
}
