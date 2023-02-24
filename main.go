package main

import (
	"log"
	"net/http"
)

func main() {

	config.Thing1a.Name = "init"

	v := Viewer{Root: config}

	log.Print("starting up...")
	http.ListenAndServe("localhost:8000", &v)

	select {}
}
