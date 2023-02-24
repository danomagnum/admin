package main

import (
	"log"
	"net/http"
)

func main() {

	config.Thing1a.Param7 = make([]int, 10)
	config.Thing1b.Param7 = make([]int, 5)

	v := Viewer{Root: config}

	log.Print("starting up...")
	http.ListenAndServe("localhost:8000", &v)

	select {}
}
