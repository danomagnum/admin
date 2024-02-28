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

	c2a := Config2{Item01: "MultiTest"}
	c2b := Config2{Item01: "AnotherTest"}
	c2c := Config2{Item01: "FinalTest"}

	c3 := Config3{Name: "MyName", Value: 42}

	v := admin.NewAdmin()
	v.RegisterStruct("Test", &config)
	v.RegisterStruct("MultiTest1", &c2a)
	v.RegisterStruct("MultiTest2", &c2b)
	v.RegisterStruct("MultiTest3", &c2c)
	c3.admin = v
	v.RegisterStruct(c3.Name, &c3)
	//v.Data["Test"] = &config

	mux.Handle("/admin/", v)

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
	Item1 string `descr:"This is a description"`
	Item2 int
	Item3 float32
	Item4 bool `descr:"I have a description too!"`
}

type Config2 struct {
	Item01 string `descr:"This is a description"`
	Item02 int
	Item03 float32
	Item04 bool `descr:"I have a description too!"`
}

func (c *Config2) Changed() {
	log.Printf("I was changed!!! %+v", *c)
}

type Config3 struct {
	Name  string
	Value int
	admin *admin.Admin
}

func (c *Config3) Change(v any) {
	n, ok := v.(*Config3)
	if !ok {
		log.Printf("I should have been a Config3 but I wasn't!!  %T", v)
	}
	if c.Name != n.Name {
		c.admin.UnRegisterStruct(c.Name)
		c.admin.RegisterStruct(n.Name, c)
		c.Name = n.Name
		log.Printf("Name change from %s to %s ", c.Name, n.Name)
	}
	if c.Value != n.Value {
		c.Value = n.Value
		log.Printf("Value change from %d to %d ", c.Value, n.Value)
	}
}
