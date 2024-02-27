package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/danomagnum/gowebstructapi"
)

func main() {

	var config Config
	config.Item1 = "initial value"
	config.Item2 = 123
	config.Item3 = 456.78
	config.Item4 = true

	mux := http.NewServeMux()

	v := gowebstructapi.NewAdmin()
	v.Data["Test"] = &config

	mux.Handle("/", v)

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
