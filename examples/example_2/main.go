package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/danomagnum/admin"
)

func main() {

	var config Config
	config.Item1 = "initial value"
	config.Item2 = 123
	config.Item3 = 456.78
	config.Item4 = true

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			f := admin.StructToForm(&config)
			out_txt := fmt.Sprintf("<html><body><form action='/' method='post'>%s<input type='submit'></form></body></html>", f)
			w.Header().Set("content-type", "text/html")
			w.Write([]byte(out_txt))
		case "POST":
			//r.ParseForm()
			//fmt.Printf("got %+v", r.Form)
			new, err := admin.RespToStruct[Config](r)
			if err != nil {
				log.Printf("got error %v\n", err)
			}
			log.Printf("got %+v", new)
			config = *new
			f := admin.StructToForm(&config)
			out_txt := fmt.Sprintf("<html><body><form action='/' method='post'>%s<input type='submit'></form></body></html>", f)
			w.Header().Set("content-type", "text/html")
			w.Write([]byte(out_txt))
		}
	})

	log.Print("starting up...")

	go func() {
		for {
			time.Sleep(3 * time.Second)
			fmt.Printf("Struct: %+v\n", config)
		}
	}()

	http.ListenAndServe("localhost:8000", mux)

	select {}

}

type Config struct {
	Item1 string
	Item2 int
	Item3 float32
	Item4 bool
}
