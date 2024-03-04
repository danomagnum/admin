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

	v := admin.NewAdmin(admin.SetDurationTimebase(time.Millisecond))
	v.RegisterStruct("Test", &config)
	v.RegisterStruct("MultiTest1", &c2a)
	v.RegisterStruct("MultiTest2", &c2b)
	v.RegisterStruct("MultiTest3", &c2c)
	v.RegisterStruct(c3.Name, &c3)

	for i := range 100 {
		c := Config3{Name: fmt.Sprintf("AutoGenStr%d", i), Value: i}
		v.RegisterStruct(c.Name, &c)
	}

	v.RegisterFunc("New Config", func() {
		v.RegisterStruct("New Config", &Config{})
	})
	v.RegisterFunc("New Config2", func() {
		v.RegisterStruct("New Config2", &Config2{})
	})
	v.RegisterFunc("New Config3", func() {
		v.RegisterStruct("New Config3", &Config3{Name: "New Config3"})
	})

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

func (c *Config2) Changed(a *admin.Admin) {
	log.Printf("I was changed!!! %+v", *c)
}

func (c *Config2) Delete(a *admin.Admin) {
	log.Printf("I was baleeted!!! %+v", *c)
}

type Config3 struct {
	Name  string
	Value int
}

func (c *Config3) Change(a *admin.Admin, v any) {
	n, ok := v.(*Config3)
	if !ok {
		log.Printf("I should have been a Config3 but I wasn't!!  %T", v)
	}
	if c.Name != n.Name {
		a.UnRegisterStruct(c.Name)
		a.RegisterStruct(n.Name, c)
		c.Name = n.Name
		log.Printf("Name change from %s to %s ", c.Name, n.Name)
	}
	if c.Value != n.Value {
		c.Value = n.Value
		log.Printf("Value change from %d to %d ", c.Value, n.Value)
	}
}
