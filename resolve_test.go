package main

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

func TestResolve(t *testing.T) {
	var c Config
	c.Thing1a.Name = "not abcd"
	c.Thing1a.Param7 = make([]int, 10)
	c.Thing1b.Param7 = make([]int, 5)

	//var c2 Config
	//c2.Thing1a.Name = "abcd"
	var c2 ConfigType1
	c2.Name = "abcd"

	//v := Viewer{Root: c}

	/*
		x, err := v.ResolvePath("")
		if err != nil {
			t.Errorf("failed to resolve path at all. %v", err)
		}
		log.Print(x.CanSet())
		log.Print(x.CanAddr())
	*/
	x := testResolvefunc(&c)

	d, _ := json.Marshal(c2)

	log.Printf("before: %+v", c)
	json.Unmarshal(d, x)
	log.Printf("after: %+v", c)

}

func testResolvefunc(t any) any {
	v := reflect.ValueOf(t).Elem()

	n := v.NumField()
	log.Printf("%d fields", n)

	f := v.FieldByName("Thing1a")

	return f.Addr().Interface()
	//return &t
}
