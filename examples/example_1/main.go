package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/danomagnum/admin"
)

func main() {

	var config Config
	config.Thing1a.Name = "init"

	v := admin.Admin{Root: config}

	log.Print("starting up...")

	go func() {
		for {
			time.Sleep(10 * time.Second)
			fmt.Printf("Struct: %+v\n", config)
		}
	}()

	http.ListenAndServe("localhost:8000", &v)

	select {}

}

type Config struct {
	Thing1a ConfigType1
	Thing1b ConfigType1
}

type Options1 int

func (o Options1) Enumerate() map[string]any {
	return map[string]any{
		"unknown": 0,
		"a":       1,
		"b":       2,
		"c":       3,
	}
}

type SubConfig struct {
	Name   string
	Length float32
	OK     bool
}

type ConfigType1 struct {
	Name   string
	Param1 int
	Param2 float32
	Param3 time.Time
	Param4 time.Duration
	Param5 bool
	Param6 Options1
	Param7 []int
	Param8 []SubConfig
}

func (t ConfigType1) Handle() string {
	return t.Name
}
func (t ConfigType1) String() string {
	return t.Name
}
func (t ConfigType1) Update(url.Values) error {
	return nil
}
func (t ConfigType1) RenderConfig() template.HTML {
	return ""
}
