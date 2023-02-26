package gowebstructapi

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
	"time"
)

type Config struct {
	Thing1a testResolveSubStr
	Thing1b testResolveSubStr
}

var config Config

type Options1 int

func (o Options1) Enumerate() map[string]any {
	return map[string]any{
		"unknown": 0,
		"a":       1,
		"b":       2,
		"c":       3,
	}
}

type testResolveSubStr2 struct {
	Name   string
	Length float32
	OK     bool
}

type testResolveSubStr struct {
	Name   string
	Param1 int
	Param2 float32
	Param3 time.Time
	Param4 time.Duration
	Param5 bool
	Param6 Options1
	Param7 []int
	Param8 []testResolveSubStr2
}

func TestResolve(t *testing.T) {
	var c Config
	c.Thing1a.Name = "not abcd"
	c.Thing1a.Param7 = make([]int, 10)
	c.Thing1b.Param7 = make([]int, 5)

	//var c2 Config
	//c2.Thing1a.Name = "abcd"
	var c2 testResolveSubStr
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
